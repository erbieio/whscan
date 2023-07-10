package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
	"server/node"
	"server/service"
)

var NotFound = fmt.Errorf("not found")

// decode parses the block
func decode(c *node.Client, ctx context.Context, number types.Long) (parsed *model.Parsed, err error) {
	// Get the block (including the transaction)
	err = c.CallContext(ctx, &parsed, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if parsed == nil {
		return nil, NotFound
	}
	// Get transaction receipt (including transaction log)
	parsed.TotalTransaction = types.Long(len(parsed.CacheTxs))
	if parsed.TotalTransaction > 0 {
		// get transaction receipt
		reqs := make([]node.BatchElem, parsed.TotalTransaction)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []any{parsed.CacheTxs[i].Hash},
				Result: &parsed.CacheTxs[i],
			}
		}
		if err = c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getTransactionReceipt err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil || parsed.CacheTxs[i].GasUsed == 0 {
				return nil, fmt.Errorf("eth_getTransactionReceipt receipt:%v,err:%v", reqs[i].Result, reqs[i].Error)
			}
		}
		// Get the receipt logs, which can only be checked according to the block hash (there may be multiple blocks with the same block height)
		err = c.CallContext(ctx, &parsed.CacheLogs, "eth_getLogs", map[string]any{"blockHash": parsed.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
	}
	// get uncle block
	if uncleCount := len(parsed.Uncles); uncleCount > 0 {
		parsed.CacheUncles = make([]*model.Uncle, uncleCount)
		reqs := make([]node.BatchElem, uncleCount)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []any{parsed.Hash, types.Long(i)},
				Result: &parsed.CacheUncles[i],
			}
		}
		if err = c.BatchCallContext(ctx, reqs); err != nil {
			return
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
		}
	}
	for _, log := range parsed.CacheLogs {
		if transferLog := utils.UnpackTransferLog(log); transferLog != nil {
			parsed.CacheTransferLogs = append(parsed.CacheTransferLogs, transferLog...)
		}
	}
	// Parse changed account properties and internal transactions
	if err = decodeInternalTxs(c, ctx, parsed); err != nil {
		return nil, fmt.Errorf("decodeInternalTxs err:%v", err)
	}
	if err = decodeAccounts(c, ctx, parsed); err != nil {
		return nil, fmt.Errorf("decodeAccounts err:%v", err)
	}
	// Parse things specific to wormholes
	err = decodeWH(c, parsed)
	tryParseMeta(parsed)
	return
}

const tracerCode = `
{
	details: [],
	enter: function (log) {
		const detail = {index:this.details.length, op:log.getType(), from:toHex(log.getFrom()), to:toHex(log.getTo()), gas:log.getGas()};
		detail.value = detail.op==='DELEGATECALL' ? '0x0' : '0x'+log.getValue().toString(16);
		this.details.push(detail);
	},
	exit: function() {},
	step: function() {},
	fault: function() {},
	result: function(ctx) {
		if (this.details.length >= 256) {return {error:ctx.error};}
		for (var i = 0; i < this.details.length; i++) {
			this.details[i].txHash = toHex(ctx.txHash);
		}
		return {details:this.details, error:ctx.error};
	}
}`

// decodeInternalTxs 获取交易的内部调用详细情况
func decodeInternalTxs(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	for _, tx := range parsed.CacheTxs {
		execRet := struct {
			Details []*model.InternalTx `json:"details"`
			Error   *string             `json:"error"`
		}{}
		if err = c.CallContext(ctx, &execRet, "debug_traceTransaction", tx.Hash, map[string]any{"tracer": tracerCode}); err != nil {
			return
		}
		tx.Error = execRet.Error
		parsed.CacheInternalTxs = append(parsed.CacheInternalTxs, execRet.Details...)
	}
	return
}

// decodeAccount to get account related properties
func decodeAccounts(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	if parsed.Number == 0 {
		state := struct {
			Accounts map[types.Address]*model.Account `json:"accounts"`
			Next     *string                          `json:"next"`
		}{}
		for next := new(string); next != nil; next, state.Next = state.Next, nil {
			if err = c.CallContext(ctx, &state, "debug_accountRange", "0x0", next, nil, false, true, true); err != nil {
				return
			}
		}
		parsed.CacheAccounts = make([]*model.Account, 0, len(state.Accounts))
		for address, account := range state.Accounts {
			account.Address = address
			account.SNFTValue = "0"
			if account.Code != nil {
				if err = utils.SetProperty(c, ctx, "0x0", account); err != nil {
					return
				}
			}
			parsed.CacheAccounts = append(parsed.CacheAccounts, account)
		}
		return
	}
	// Get the change account address
	var modifiedAccounts []types.Address
	if err = c.CallContext(ctx, &modifiedAccounts, "debug_getModifiedAccountsByHash", parsed.Hash); err != nil {
		return
	}
	if len(modifiedAccounts) > 0 {
		parsed.CacheAccounts = make([]*model.Account, 0, len(modifiedAccounts))
		contracts := make(map[types.Address]*model.Account)
		for _, tx := range parsed.CacheTxs {
			if address := tx.ContractAddress; address != nil {
				contracts[*address] = &model.Account{Creator: &tx.From, CreatedTx: &tx.Hash}
			}
		}
		for _, tx := range parsed.CacheInternalTxs {
			if tx.Op == "CREATE" || tx.Op == "CREATE2" {
				contracts[tx.To] = &model.Account{Creator: &tx.From, CreatedTx: &tx.TxHash}
			}
		}
		number := parsed.Number.Hex()
		info := struct {
			Nonce   types.Long `json:"Nonce"`
			Balance *big.Int   `json:"Balance"`
			Worm    *struct {
				VoteWeight *big.Int `json:"VoteWeight"`
			} `json:"Worm"`
		}{}
		for _, address := range modifiedAccounts {
			if address != types.ZeroAddress && (address[:12] == "0x0000000000" || address[:12] == "0x8000000000") {
				continue
			}
			account := &model.Account{Address: address, Number: parsed.Number, SNFTValue: "0"}
			if contract := contracts[address]; contract != nil {
				account.Creator, account.CreatedTx = contract.Creator, contract.CreatedTx
				if err = utils.SetProperty(c, ctx, number, account); err != nil {
					return
				}
				if err = c.CallContext(ctx, &account.Code, "eth_getCode", address, number); err != nil {
					return
				}

			}
			if err = c.CallContext(ctx, &info, "eth_getAccountInfo", address, number); err != nil {
				return
			}
			account.Number = parsed.Number
			account.Nonce = info.Nonce
			account.Balance = types.BigInt(info.Balance.String())
			if info.Worm != nil {
				account.SNFTValue = info.Worm.VoteWeight.String()
			} else {
				account.SNFTValue = "0"
			}
			parsed.CacheAccounts = append(parsed.CacheAccounts, account)
		}
	}
	return
}

func write(c *node.Client, ctx context.Context, parsed *model.Parsed) (head types.Long, err error) {
	head, err = service.Insert(parsed)
	if err != nil || parsed.Number == head {
		return
	}
	for parsed.Number = parsed.Number - 2; parsed.Number >= 0; parsed.Number-- {
		number, pass := parsed.Number.Hex(), false
		if err = c.CallContext(ctx, &parsed.Block, "eth_getBlockByNumber", number, false); err != nil {
			return
		}
		if pass, err = service.VerifyHead(parsed); err != nil {
			return
		} else if pass {
			for _, account := range parsed.CacheAccounts {
				info := struct {
					Nonce   types.Long `json:"Nonce"`
					Balance *big.Int   `json:"Balance"`
					Worm    *struct {
						VoteWeight *big.Int `json:"VoteWeight"`
					} `json:"Worm"`
				}{}
				if err = c.Call(&info, "eth_getAccountInfo", account.Address, number); err != nil {
					return
				}
				account.Number = parsed.Number
				account.Nonce = info.Nonce
				account.Balance = types.BigInt(info.Balance.String())
				if info.Worm != nil {
					account.SNFTValue = info.Worm.VoteWeight.String()
				} else {
					account.SNFTValue = "0"
				}
			}
			break
		}
	}
	return parsed.Number, service.SetHead(parsed)
}

func check(c *node.Client, ctx context.Context) (stats *model.Stats, err error) {
	if err = c.CallContext(ctx, &struct{}{}, "debug_gcStats"); err != nil {
		return
	}
	if err = c.CallContext(ctx, &struct{}{}, "eth_getAccountInfo", types.ZeroAddress, "0x0"); err != nil {
		return
	}
	chainId, genesis, stats := types.Long(0), model.Header{}, service.GetStats()
	if err = c.CallContext(ctx, &chainId, "eth_chainId"); err != nil {
		return
	}
	if err = c.CallContext(ctx, &genesis, "eth_getBlockByNumber", "0x0", false); err != nil {
		return
	}
	if stats.TotalBlock > 0 && (stats.ChainId != int64(chainId) || stats.Genesis != genesis) {
		err = errors.New("stored data and chain node information do not match")
	} else {
		stats.ChainId = int64(chainId)
		stats.Genesis = genesis
	}
	return
}

// decodeWH Imports the underlying NFT transaction of the SNFT meta information distributed by the block
func decodeWH(c *node.Client, wh *model.Parsed) (err error) {
	if number := wh.Number.Hex(); wh.Number > 0 {
		// Miner reward SNFT processing
		var rewards []*struct {
			Address      string   `json:"Address"`
			NFTAddress   *string  `json:"NftAddress"`
			RewardAmount *big.Int `json:"RewardAmount"`
		}
		err = c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", number, true)
		if err != nil {
			return fmt.Errorf("GetReward() err:%v", err)
		}
		for i := range rewards {
			identity := uint8(0)
			switch i {
			case 7, 8, 9, 10:
				identity = 3
			case 0, 1, 2, 3, 4, 5:
				identity = 2
			case 6:
				identity = 1
			default:
				return fmt.Errorf("reward length more than 11")
			}
			wh.Rewards = append(wh.Rewards, &model.Reward{
				Address:     rewards[i].Address,
				Identity:    identity,
				BlockNumber: int64(wh.Number),
			})
			if rewards[i].RewardAmount == nil {
				// Note that when NFTAddress is zero address error
				wh.Rewards[i].SNFT = rewards[i].NFTAddress
				// Parse the new phase ID
				if addr := *rewards[i].NFTAddress; addr[39:] == "000" {
					// Write the current information, once every 4096 SNFT rewards
					epoch := struct {
						StartIndex int64    `json:"start_Index"`
						Dir        string   `json:"dir"`
						Royalty    int64    `json:"royalty"`
						Address    string   `json:"Address"`
						Creator    string   `json:"creator"`
						VoteWeight *big.Int `json:"vote_weight"`
					}{}
					if err = c.Call(&epoch, "eth_getCurrentNFTInfo", number); err != nil {
						return fmt.Errorf("GetEpoch() err:%v", err)
					}
					if len(epoch.Dir) == 52 {
						epoch.Dir = epoch.Dir + "/"
					}
					wh.Epoch = &model.Epoch{
						ID:           addr[:39],
						Creator:      strings.ToLower(epoch.Creator),
						RoyaltyRatio: epoch.Royalty,
						Dir:          epoch.Dir,
						WeightValue:  epoch.VoteWeight.Text(10),
						Voter:        strings.ToLower(epoch.Address),
						StartNumber:  int64(wh.Number),
						StartTime:    int64(wh.Timestamp),
					}

					selected := wh.Number - 1
					for startIndex := epoch.StartIndex; selected > wh.Number-64 && selected > 0; selected-- {
						if err = c.Call(&epoch, "eth_getCurrentNFTInfo", selected.Hex()); err != nil {
							return fmt.Errorf("GetEpoch() err:%v", err)
						}
						if startIndex != epoch.StartIndex {
							break
						}
					}
					info := struct {
						Balance *big.Int `json:"Balance"`
					}{}
					if err = c.Call(&info, "eth_getAccountInfo", "0xffffffffffffffffffffffffffffffffffffffff", selected.Hex()); err != nil {
						return
					}
					wh.Epoch.Number = int64(selected + 1)
					wh.Epoch.Reward = info.Balance.Text(10)
					wh.Epoch.Profit = "0"
				}
			} else {
				wh.Rewards[i].Amount = new(string)
				*wh.Rewards[i].Amount = rewards[i].RewardAmount.Text(10)
			}
		}

		var onlineWeight []*struct {
			Address string `json:"address"`
			Value   int64  `json:"value"`
		}
		if err = c.Call(&onlineWeight, "eth_getValidators", number); err != nil {
			return fmt.Errorf("getWeights() err:%v", err)
		}
		wh.ChangeValidators = make([]*model.Validator, 0, len(onlineWeight))
		for _, weight := range onlineWeight {
			wh.ChangeValidators = append(wh.ChangeValidators, &model.Validator{Address: weight.Address, Amount: "0", Weight: weight.Value})
		}

		if wh.Miner == types.ZeroAddress {
			wh.Validators = make([]*model.Penalty, 0, len(onlineWeight))
			for _, weight := range onlineWeight {
				wh.Validators = append(wh.Validators, &model.Penalty{Address: types.Address(weight.Address), Weight: types.Long(weight.Value)})
			}
			var proposers []*struct {
				Address string
			}
			if err = c.Call(&proposers, "eth_getRealParticipantsByNumber", number); err != nil {
				return fmt.Errorf("getProposers() err:%v", err)
			}
			for _, proposer := range proposers {
				wh.Proposers = append(wh.Proposers, types.Address(proposer.Address))
			}
		}

		// wormholes transaction processing
		for _, tx := range wh.CacheTxs {
			err = decodeWHTx(wh, tx)
			if err != nil {
				return err
			}
		}

		// wormholes auto merge snft
		for _, eventLog := range wh.CacheLogs {
			if len(eventLog.Topics) == 3 && len(eventLog.Data) >= 66 {
				if eventLog.Topics[0] == "0x77415a68a0d28daf11e1308e53371f573e0920810c9cd9de7904777d5fb9d625" {
					pieces, _ := strconv.ParseInt(eventLog.Data[62:66], 16, 32)
					if pieces > 0 {
						addr := string(eventLog.Topics[1][27:])
						for i := 0; i < 3; i++ {
							if addr[i] == '8' {
								wh.Mergers = append(wh.Mergers, &model.SNFT{
									Address:      "0x" + addr[i:],
									TxAmount:     "0",
									RewardAt:     int64(wh.Timestamp),
									RewardNumber: int64(wh.Number),
									Owner:        "0x" + string(eventLog.Topics[2][26:]),
									Pieces:       pieces,
								})
							}
						}
					}
				}
			}
		}
		for _, reward := range wh.Rewards {
			if reward.SNFT != nil && (*reward.SNFT)[41] == 'f' {
				addr := (*reward.SNFT)[:41] + "0"
				for i := 0; i < 3; i++ {
					info := struct {
						NFT *struct {
							MergeLevel  int    `json:"MergeLevel"`
							MergeNumber int64  `json:"MergeNumber"`
							Owner       string `json:"Owner"`
						} `json:"Nft"`
					}{}
					if err = c.Call(&info, "eth_getAccountInfo", addr, number); err != nil {
						return
					}
					if info.NFT != nil && info.NFT.MergeLevel > i {
						wh.Mergers = append(wh.Mergers, &model.SNFT{
							Address:      addr[:41-i],
							TxAmount:     "0",
							RewardAt:     int64(wh.Timestamp),
							RewardNumber: int64(wh.Number),
							Owner:        info.NFT.Owner,
							Pieces:       info.NFT.MergeNumber,
						})
						addr = addr[:40-i] + "0" + addr[41+i:]
					} else {
						break
					}
				}
			}
		}
	} else {
		info := struct {
			Worm *struct {
				ExchangerBalance *big.Int `json:"ExchangerBalance"`
				FeeRate          int64    `json:"FeeRate"`
				ExchangerName    string   `json:"ExchangerName"`
				ExchangerURL     string   `json:"ExchangerURL"`
			} `json:"Worm"`
		}{}
		for _, account := range wh.CacheAccounts {
			if err = c.Call(&info, "eth_getAccountInfo", account.Address, "0x0"); err != nil {
				return
			}
			if info.Worm != nil && info.Worm.ExchangerBalance.Int64() != 0 {
				balance := info.Worm.ExchangerBalance.Text(10)
				wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
					Address:   string(account.Address),
					Name:      info.Worm.ExchangerName,
					URL:       info.Worm.ExchangerURL,
					FeeRatio:  info.Worm.FeeRate,
					Creator:   string(account.Address),
					Timestamp: int64(wh.Timestamp),
					TxHash:    "0x0",
					Amount:    balance,
				})
			}
		}
		result := struct {
			Validators []*struct {
				Addr    string   `json:"Addr"`
				Balance *big.Int `json:"Balance"`
				Proxy   string   `json:"Proxy"`
			} `json:"Validators"`
		}{}
		if err = c.Call(&result, "eth_getValidator", "0x0"); err != nil {
			return
		}
		for _, validator := range result.Validators {
			wh.ChangeValidators = append(wh.ChangeValidators, &model.Validator{
				Address: validator.Addr,
				Amount:  validator.Balance.Text(10),
				Proxy:   validator.Proxy,
			})
		}
	}
	return
}

// decodeWHTx parses the special transaction of the wormholes blockchain
func decodeWHTx(wh *model.Parsed, tx *model.Transaction) (err error) {
	input, _ := hex.DecodeString(tx.Input[2:])
	// Non-wormholes and failed transactions are not resolved
	if len(input) < 10 || string(input[0:10]) != "wormholes:" || *tx.Status == 0 {
		return
	}
	type Wormholes struct {
		Type         uint8  `json:"type"`
		NFTAddress   string `json:"nft_address"`
		ProxyAddress string `json:"proxy_address"`
		Exchanger    string `json:"exchanger"`
		Royalty      int64  `json:"royalty"`
		MetaURL      string `json:"meta_url"`
		FeeRate      int64  `json:"fee_rate"`
		Name         string `json:"name"`
		Url          string `json:"url"`
		Dir          string `json:"dir"`
		StartIndex   string `json:"start_index"`
		Number       uint64 `json:"number"`
		Buyer        struct {
			Amount      string `json:"price"`
			NFTAddress  string `json:"nft_address"`
			Exchanger   string `json:"exchanger"`
			BlockNumber string `json:"block_number"`
			Seller      string `json:"seller"`
			Sig         string `json:"sig"`
		} `json:"buyer"`
		Seller1 struct {
			Amount      string `json:"price"`
			NFTAddress  string `json:"nft_address"`
			Exchanger   string `json:"exchanger"`
			BlockNumber string `json:"block_number"`
			Seller      string `json:"seller"`
			Sig         string `json:"sig"`
		} `json:"seller1"`
		Seller2 struct {
			Amount        string `json:"price"`
			Royalty       string `json:"royalty"`
			MetaURL       string `json:"meta_url"`
			ExclusiveFlag string `json:"exclusive_flag"`
			Exchanger     string `json:"exchanger"`
			BlockNumber   string `json:"block_number"`
			Sig           string `json:"sig"`
		} `json:"seller2"`
		ExchangerAuth struct {
			ExchangerOwner string `json:"exchanger_owner"`
			To             string `json:"to"`
			BlockNumber    string `json:"block_number"`
			Sig            string `json:"sig"`
		} `json:"exchanger_auth"`
		Creator    string `json:"creator"`
		Version    string `json:"version"`
		RewardFlag uint8  `json:"reward_flag"`
	}
	var w Wormholes
	if err = json.Unmarshal(input[10:], &w); err != nil {
		return
	}

	blockNumber := int64(wh.Number)
	timestamp := int64(wh.Timestamp)
	txHash := string(tx.Hash)
	from := string(tx.From)
	value := string(tx.Value)
	switch w.Type {
	case 0: //Users mint NFT by themselves
		nftAddr := "" //Calculate fill in real time when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  w.Royalty, //The unit is one ten thousandth
			MetaUrl:       realMeatUrl(w.MetaURL),
			RawMetaUrl:    w.MetaURL,
			ExchangerAddr: strings.ToLower(w.Exchanger),
			TxAmount:      "0",
			Creator:       string(*tx.To),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(*tx.To),
		})

	case 1: //Users transfer NFT by themselves
		w.NFTAddress = strings.ToLower(w.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:      1,
			NFTAddr:     &w.NFTAddress,
			To:          string(*tx.To),
			Price:       value,
			Timestamp:   timestamp,
			TxHash:      txHash,
			BlockNumber: blockNumber,
		})

	case 6: //recycle snft
		w.NFTAddress = strings.ToLower(w.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:      6,
			NFTAddr:     &w.NFTAddress,
			From:        from,
			Timestamp:   timestamp,
			TxHash:      txHash,
			BlockNumber: blockNumber,
		})

	case 9, 10: //validator pledge, can be pledged multiple times, starting at 100000ERB
		validator := &model.Validator{Address: from, Amount: value}
		if w.Type == 10 && value != "0" {
			validator.Amount = "-" + value
		}
		if len(w.ProxyAddress) == 42 && w.ProxyAddress != types.ZeroAddress {
			validator.Proxy = w.ProxyAddress
		}
		wh.ChangeValidators = append(wh.ChangeValidators, validator)
		wh.Pledges = append(wh.Pledges, &model.Pledge{
			Address:   from,
			Type:      int64(w.Type),
			Amount:    value,
			Number:    blockNumber,
			Timestamp: timestamp,
		})

	case 11: //Open the exchange
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
			Address:     from,
			Name:        w.Name,
			URL:         w.Url,
			FeeRatio:    w.FeeRate,
			Creator:     from,
			Timestamp:   timestamp,
			BlockNumber: blockNumber,
			TxHash:      txHash,
			Amount:      value,
		})

	case 12: //Close the exchange
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{Address: from, Amount: "0", CloseAt: &timestamp})

	case 14: //NFT bid transaction (initiated by the seller or the exchange, and the buyer signs the price)
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		w.Buyer.Exchanger = strings.ToLower(w.Buyer.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        14,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &w.Buyer.Exchanger,
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 15: //NFT pricing purchase transaction (buyer initiates, seller signs price)
		w.Seller1.NFTAddress = strings.ToLower(w.Seller1.NFTAddress)
		w.Seller1.Exchanger = strings.ToLower(w.Seller1.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        15,
			NFTAddr:       &w.Seller1.NFTAddress,
			ExchangerAddr: &w.Seller1.Exchanger,
			To:            from,  //The transaction initiator is the buyer
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 16: //NFT lazy pricing purchase transaction, the buyer initiates (the NFT is minted first, and the seller signs the price)
		// Restore NFT creator address (also seller address) from signature
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// royalty rate string to number
		royaltyRatio, err := strconv.ParseInt(w.Seller2.Royalty[2:], 16, 64)
		if err != nil {
			return err
		}
		w.Seller2.Exchanger = strings.ToLower(w.Seller2.Exchanger)
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  royaltyRatio,
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: w.Seller2.Exchanger,
			TxAmount:      "0",
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        16,
			NFTAddr:       &nftAddr,
			ExchangerAddr: &w.Seller2.Exchanger,
			From:          string(creator),
			To:            from,  //The transaction initiator is the buyer
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 17: //NFT lazy pricing purchase transaction, initiated by the exchange (mint NFT first, and the seller signs the price)
		// Restore NFT creator address (also seller address) from signature
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// royalty rate string to number
		royaltyRatio, err := strconv.ParseInt(w.Seller2.Royalty[2:], 16, 64)
		if err != nil {
			return err
		}
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  royaltyRatio,
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: from, //The transaction initiator is the exchange address
			TxAmount:      "0",
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        17,
			NFTAddr:       &nftAddr,
			ExchangerAddr: &from, //The transaction initiator is the exchange address
			From:          string(creator),
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 18: //The NFT bid transaction is initiated by the address authorized by the exchange (the buyer signs the price)
		// restore the exchange address from the authorized signature
		msg := w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := utils.RecoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return err
		}
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        18,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: (*string)(&exchangerAddr),
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 19: //NFT lazy bid transaction, initiated by the address authorized by the exchange (the buyer signs the price)
		// Restore NFT creator address (also seller address) from signature
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// restore the exchange address from the authorized signature
		msg = w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := utils.RecoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return err
		}
		// royalty rate string to number
		royaltyRatio, err := strconv.ParseInt(w.Seller2.Royalty[2:], 16, 64)
		if err != nil {
			return err
		}
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  royaltyRatio,
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: string(exchangerAddr), //The transaction initiator is the exchange address
			TxAmount:      "0",
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        19,
			NFTAddr:       &nftAddr,
			ExchangerAddr: (*string)(&exchangerAddr),
			From:          string(creator),
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 20: //NFT matches the transaction, and the exchange initiates it
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        20,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &from,
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 21: //Exchange pledge
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
			Address: from,
			Amount:  value,
		})

	case 22: //Revoke the exchange pledge
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
			Address: from,
			Amount:  "-" + value,
		})
	case 27: //forcibly buy snft that does not belong to you(level 1 address)
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		w.Buyer.Exchanger = strings.ToLower(w.Buyer.Exchanger)
		w.Buyer.Seller = strings.ToLower(w.Buyer.Seller)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        27,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &w.Buyer.Exchanger,
			From:          w.Buyer.Seller,
			To:            string(*tx.To),
			Price:         value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})
	case 28: //forcibly buy snft that does not belong to you(level 1 address)
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		w.Buyer.Exchanger = strings.ToLower(w.Buyer.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        28,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &w.Buyer.Exchanger,
			From:          types.ZeroAddress,
			To:            string(*tx.To),
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})
	case 31:
		wh.ChangeValidators = append(wh.ChangeValidators, &model.Validator{Address: from, Proxy: w.ProxyAddress, Amount: "0"})
	}
	return
}

// realMeatUrl parses the real metaUrl
func realMeatUrl(meta string) string {
	data, err := hex.DecodeString(meta)
	if err != nil {
		return ""
	}
	r := struct {
		Meta string `json:"meta"`
	}{}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return ""
	}
	return r.Meta
}

// tryParseMeta parses NFT AND SNFT meta information
func tryParseMeta(wh *model.Parsed) {
	for _, nft := range wh.NFTs {
		nftMeta, err := utils.GetNFTMeta(nft.MetaUrl)
		if err != nil {
			if strings.Index(err.Error(), "context deadline exceeded") > 0 {
				nft.Status += 1
			} else {
				nft.Status = -1
			}
			log.Println("Failed to parse NFT meta information", nft.Address, nft.MetaUrl, err)
			continue
		}

		//collection name + collection creator + hash of the exchange where the collection is located
		if nftMeta.CollectionsName != "" && nftMeta.CollectionsCreator != "" {
			nft.CollectionId = string(utils.Keccak256Hash(
				[]byte(nftMeta.CollectionsName),
				[]byte(nftMeta.CollectionsCreator),
				[]byte(nftMeta.CollectionsExchanger),
			))
			wh.Collections = append(wh.Collections, &model.Collection{
				Id:          nft.CollectionId,
				Name:        nftMeta.CollectionsName,
				Creator:     nftMeta.CollectionsCreator,
				Category:    nftMeta.CollectionsCategory,
				Desc:        nftMeta.CollectionsDesc,
				ImgUrl:      nftMeta.CollectionsImgUrl,
				BlockNumber: int64(wh.Number),
				Exchanger:   &nftMeta.CollectionsExchanger,
			})
		}
		nft.Name = nftMeta.Name
		nft.Desc = nftMeta.Desc
		nft.Attributes = nftMeta.Attributes
		nft.Category = nftMeta.Category
		nft.SourceUrl = nftMeta.SourceUrl
	}

	if epoch := wh.Epoch; epoch != nil {
		for i := 0; i < 16; i++ {
			hexI := fmt.Sprintf("%x", i)
			collectionId := epoch.ID + hexI
			metaUrl := ""
			if epoch.Dir != "" {
				metaUrl = epoch.Dir + hexI + "0"
			}
			// write collection information
			collection := &model.Collection{Id: collectionId, MetaUrl: metaUrl, BlockNumber: epoch.Number}
			if metaUrl != "" {
				nftMeta, err := utils.GetNFTMeta(metaUrl)
				if err != nil {
					log.Println("Failed to parse SNFT collection information", collectionId, metaUrl, err)
				} else {
					collection.Name = nftMeta.CollectionsName
					collection.Desc = nftMeta.CollectionsDesc
					collection.Category = nftMeta.CollectionsCategory
					collection.ImgUrl = nftMeta.CollectionsImgUrl
					collection.Creator = nftMeta.CollectionsCreator
				}
			}
			wh.Collections = append(wh.Collections, collection)
			for j := 0; j < 16; j++ {
				hexJ := fmt.Sprintf("%x", j)
				FNFTId := collectionId + hexJ
				if epoch.Dir != "" {
					metaUrl = epoch.Dir + hexI + hexJ
				}
				// write complete SNFT information
				fnft := &model.FNFT{ID: FNFTId, MetaUrl: metaUrl}
				if metaUrl != "" {
					nftMeta, err := utils.GetNFTMeta(metaUrl)
					if err != nil {
						if strings.Index(err.Error(), "context deadline exceeded") > 0 {
							fnft.Status += 1
						} else {
							fnft.Status = -1
						}
						log.Println("Failed to parse and store SNFT meta information", FNFTId, metaUrl, err)
					} else {
						fnft.Name = nftMeta.Name
						fnft.Desc = nftMeta.Desc
						fnft.Attributes = nftMeta.Attributes
						fnft.Category = nftMeta.Category
						fnft.SourceUrl = nftMeta.SourceUrl
					}
				}
				wh.FNFTs = append(wh.FNFTs, fnft)
			}
		}
	}
}
