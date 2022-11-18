package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
func decode(c *node.Client, ctx context.Context, number types.Uint64) (*model.Parsed, error) {
	var raw json.RawMessage
	// Get the block (including the transaction)
	err := c.CallContext(ctx, &raw, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if len(raw) == 0 {
		return nil, NotFound
	}
	var parsed model.Parsed
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	}
	// Get transaction receipt (including transaction log)
	if parsed.TotalTransaction = types.Uint64(len(parsed.CacheTxs)); parsed.TotalTransaction > 0 {
		// get transaction receipt
		reqs := make([]node.BatchElem, parsed.TotalTransaction)
		for i, tx := range parsed.CacheTxs {
			reqs[i] = node.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{tx.Hash},
				Result: &parsed.CacheTxs[i],
			}
		}
		if err := c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getTransactionReceipt err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil || parsed.CacheTxs[i].GasUsed == 0 {
				return nil, fmt.Errorf("eth_getTransactionReceipt receipt:%v,err:%v", reqs[i].Result, reqs[i].Error)
			}
		}
		// Get the receipt logs, which can only be checked according to the block hash (there may be multiple blocks with the same block height)
		err := c.CallContext(ctx, &parsed.CacheLogs, "eth_getLogs", map[string]interface{}{"blockHash": parsed.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
	}
	// get uncle block
	if parsed.UnclesCount = types.Uint64(len(parsed.UncleHashes)); parsed.UnclesCount > 0 {
		parsed.CacheUncles = make([]*model.Uncle, parsed.UnclesCount)
		reqs := make([]node.BatchElem, parsed.UnclesCount)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{parsed.Hash, types.Uint64(i)},
				Result: &parsed.CacheUncles[i],
			}
		}
		if err := c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getUncleByBlockHashAndIndex err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, fmt.Errorf("eth_getUncleByBlockHashAndIndex err:%v", reqs[i].Error)
			}
		}
	}
	for _, log := range parsed.CacheLogs {
		if transferLog := utils.UnpackTransferLog(log); transferLog != nil {
			parsed.CacheTransferLogs = append(parsed.CacheTransferLogs, transferLog...)
		}
	}
	// Parse changed account properties and internal transactions
	err = decodeAccounts(c, ctx, &parsed)
	if err != nil {
		return nil, fmt.Errorf("decodeAccounts err:%v", err)
	}
	err = decodeInternalTxs(c, ctx, &parsed)
	if err != nil {
		return nil, fmt.Errorf("decodeInternalTxs err:%v", err)
	}
	// Parse things specific to wormholes
	err = decodeWH(c, &parsed)

	return &parsed, err
}

type ExecutionResult struct {
	Failed      bool   `json:"failed"`
	ReturnValue string `json:"returnValue"`
	StructLogs  []struct {
		Op      string   `json:"op"`
		Gas     uint64   `json:"gas"`
		GasCost uint64   `json:"gasCost"`
		Depth   int      `json:"depth"`
		Error   string   `json:"error,omitempty"`
		Stack   []string `json:"stack,omitempty"`
	} `json:"structLogs"`
}

var params = struct {
	DisableStorage bool `json:"disableStorage"`
	DisableMemory  bool `json:"disableMemory"`
	Limit          int  `json:"limit"`
}{true, true, 81920}

func decodeInternalTxs(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	for _, tx := range parsed.CacheTxs {
		var execRet ExecutionResult
		if err = c.CallContext(ctx, &execRet, "debug_traceTransaction", tx.Hash, params); err != nil {
			return
		}
		if len(execRet.StructLogs) >= params.Limit {
			continue
		}
		caller := tx.To
		if caller == nil {
			caller = tx.ContractAddress
		}
		callers := []*types.Address{caller}
		iTx := &model.InternalTx{TxHash: tx.Hash, BlockNumber: parsed.Number, To: new(types.Address), Value: "0x0"}
		for _, log := range execRet.StructLogs {
			if len(*caller) == 0 {
				*caller = *utils.HexToAddress(log.Stack[len(log.Stack)-1])
			}
			switch log.Op {
			case "CALL", "CALLCODE":
				iTx.To = utils.HexToAddress(log.Stack[len(log.Stack)-2])
				iTx.Value = utils.HexToBigInt(log.Stack[len(log.Stack)-3][2:])
				callers = append(callers, iTx.To)
			case "DELEGATECALL":
				iTx.To = utils.HexToAddress(log.Stack[len(log.Stack)-2])
				callers = append(callers, callers[log.Depth-1])
			case "STATICCALL":
				iTx.To = utils.HexToAddress(log.Stack[len(log.Stack)-2])
				callers = append(callers, iTx.To)
			case "CREATE", "CREATE2":
				iTx.Value = utils.HexToBigInt(log.Stack[len(log.Stack)-1][2:])
				callers = append(callers, iTx.To)
			case "SELFDESTRUCT":
				iTx.To = utils.HexToAddress(log.Stack[len(log.Stack)-1])
				caller = callers[len(callers)-1]
			case "RETURN", "STOP", "REVERT":
				caller = callers[len(callers)-1]
				callers = callers[:len(callers)-1]
				continue
			default:
				continue
			}
			iTx.Depth = types.Uint64(log.Depth)
			iTx.Op = log.Op
			iTx.From = callers[log.Depth-1]
			iTx.GasLimit = types.Uint64(log.Gas)
			parsed.CacheInternalTxs = append(parsed.CacheInternalTxs, iTx)
			iTx = &model.InternalTx{TxHash: tx.Hash, BlockNumber: parsed.Number, To: new(types.Address), Value: "0x0"}
		}
	}
	return nil
}

// decodeAccount to get account related properties
func decodeAccounts(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	if parsed.Number == 0 {
		return decodeState(c, ctx, parsed)
	}
	// Get the change account address
	var modifiedAccounts []types.Address
	if err = c.CallContext(ctx, &modifiedAccounts, "debug_getModifiedAccountsByHash", parsed.Hash); err != nil {
		return
	}
	parsed.CacheAccounts = make(map[types.Address]*model.Account)
	for _, address := range modifiedAccounts {
		if address[:12] != "0x0000000000" && address[:12] != "0x8000000000" {
			parsed.CacheAccounts[address] = &model.Account{Address: address}
		}
	}
	if len(parsed.CacheAccounts) > 0 {
		reqs := make([]node.BatchElem, 0, 2*len(parsed.CacheAccounts))
		for _, account := range parsed.CacheAccounts {
			reqs = append(reqs, node.BatchElem{
				Method: "eth_getBalance",
				Args:   []interface{}{account.Address, parsed.Number.Hex()},
				Result: &account.Balance,
			})
			reqs = append(reqs, node.BatchElem{
				Method: "eth_getTransactionCount",
				Args:   []interface{}{account.Address, parsed.Number.Hex()},
				Result: &account.Nonce,
			})
		}
		if err = c.BatchCallContext(ctx, reqs); err != nil {
			return
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return reqs[i].Error
			}
		}

		var contracts []types.Address
		for _, tx := range parsed.CacheTxs {
			if tx.ContractAddress != nil && parsed.CacheAccounts[*tx.ContractAddress] != nil {
				contracts = append(contracts, *tx.ContractAddress)
				parsed.CacheAccounts[*tx.ContractAddress].Creator = &tx.From
				parsed.CacheAccounts[*tx.ContractAddress].CreatedTx = &tx.Hash
			}
		}
		for _, tx := range parsed.CacheInternalTxs {
			if (tx.Op == "CREATE" || tx.Op == "CREATE2") && parsed.CacheAccounts[*tx.To] != nil {
				contracts = append(contracts, *tx.To)
				parsed.CacheAccounts[*tx.To].Creator = tx.From
				parsed.CacheAccounts[*tx.To].CreatedTx = &tx.TxHash
			}
		}
		if len(contracts) > 0 {
			reqs := make([]node.BatchElem, 0, len(contracts))
			for _, contract := range contracts {
				reqs = append(reqs, node.BatchElem{
					Method: "eth_getCode",
					Args:   []interface{}{contract, parsed.Number.Hex()},
					Result: &parsed.CacheAccounts[contract].Code,
				})
			}
			if err = c.BatchCallContext(ctx, reqs); err != nil {
				return
			}
			for i := range reqs {
				if reqs[i].Error != nil {
					return reqs[i].Error
				}
			}
			for _, contract := range contracts {
				account := parsed.CacheAccounts[contract]
				account.Name, account.Symbol, account.Type, err = utils.Property(c, contract)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func decodeState(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	status := struct {
		Accounts map[types.Address]*model.Account `json:"accounts"`
		Next     *string                          `json:"next"`
	}{}
	for next, number := new(string), parsed.Number.Hex(); next != nil; next, status.Next = status.Next, nil {
		if err = c.CallContext(ctx, &status, "debug_accountRange", number, next, nil, true, true, true); err != nil {
			return
		}
	}
	for address, account := range status.Accounts {
		if address[:12] != "0x0000000000" && address[:12] != "0x8000000000" {
			account.Address = address
		} else {
			delete(status.Accounts, address)
		}
	}
	parsed.CacheAccounts = status.Accounts
	return
}

func write(c *node.Client, ctx context.Context, parsed *model.Parsed) (types.Uint64, error) {
	badBlocks, err := service.Insert(parsed)
	if err != nil || badBlocks == nil {
		return parsed.Number, err
	}
	for _, block := range badBlocks {
		parsed.Number--
		var result *struct{ Hash types.Hash }
		if err = c.CallContext(ctx, &result, "eth_getBlockByNumber", parsed.Number.Hex(), true); err != nil {
			return 0, err
		} else if result != nil && result.Hash == block {
			break
		}
	}
	if err = decodeState(c, ctx, parsed); err != nil {
		return 0, err
	}
	return parsed.Number, service.SetHead(parsed)
}

func check(c *node.Client, ctx context.Context) (stats *model.Stats, err error) {
	if err = c.CallContext(ctx, &struct{}{}, "debug_gcStats"); err != nil {
		return
	}
	if err = c.CallContext(ctx, &struct{}{}, "eth_getAccountInfo", "0x0000000000000000000000000000000000000000", "0x0"); err != nil {
		return
	}
	chainId, genesis, stats := types.Uint64(0), model.Header{}, service.GetStats()
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
func decodeWH(c *node.Client, wh *model.Parsed) error {
	if wh.Number > 0 {
		// Miner reward SNFT processing
		var rewards []*struct {
			Address      string   `json:"Address"`
			NFTAddress   *string  `json:"NftAddress"`
			RewardAmount *big.Int `json:"RewardAmount"`
		}
		err := c.Call(&rewards, "eth_getBlockBeneficiaryAddressByNumber", wh.Number.Hex(), true)
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
				BlockNumber: uint64(wh.Number),
			})
			if rewards[i].RewardAmount == nil {
				// Note that when NFTAddress is zero address error
				wh.Rewards[i].SNFT = rewards[i].NFTAddress
				nftAddr := *rewards[i].NFTAddress
				// Parse the new phase ID
				addr, _ := new(big.Int).SetString(nftAddr[3:], 16)
				if addr.Mod(addr, big.NewInt(4096)).Uint64() == 0 {
					epochId := nftAddr[:39]
					// Write the current information, once every 4096 SNFT rewards
					if len(epochId) > 0 {
						epoch := struct {
							Dir        string   `json:"dir"`
							Royalty    uint32   `json:"royalty"`
							Creator    string   `json:"creator"`
							Address    string   `json:"address"` //Exchange address
							VoteWeight *big.Int `json:"vote_weight"`
						}{}
						if err = c.Call(&epoch, "eth_getCurrentNFTInfo", wh.Number.Hex()); err != nil {
							return fmt.Errorf("GetEpoch() err:%v", err)
						}
						if len(epoch.Dir) == 52 {
							epoch.Dir = epoch.Dir + "/"
						}
						wh.Epoch = &model.Epoch{
							ID:           epochId,
							Creator:      strings.ToLower(epoch.Creator),
							RoyaltyRatio: epoch.Royalty,
							Dir:          epoch.Dir,
							Exchanger:    epoch.Address,
							VoteWeight:   epoch.VoteWeight.Text(10),
							Number:       uint64(wh.Number),
							Timestamp:    uint64(wh.Timestamp),
						}
					}
				}
			} else {
				wh.Rewards[i].Amount = new(string)
				*wh.Rewards[i].Amount = rewards[i].RewardAmount.Text(10)
			}
		}

		// wormholes transaction processing
		for _, tx := range wh.CacheTxs {
			err = decodeWHTx(c, wh, tx)
			if err != nil {
				return err
			}
		}
	} else {
		info := struct {
			ExchangerBalance *big.Int `json:"ExchangerBalance"`
			FeeRate          uint32   `json:"FeeRate"`
			ExchangerName    string   `json:"ExchangerName"`
			ExchangerURL     string   `json:"ExchangerURL"`
		}{}
		for address := range wh.CacheAccounts {
			if err := c.Call(&info, "eth_getAccountInfo", address, "0x0"); err != nil {
				return err
			}
			if balance := info.ExchangerBalance.Text(10); balance != "0" {
				wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
					Address:   string(address),
					Name:      info.ExchangerName,
					URL:       info.ExchangerURL,
					FeeRatio:  info.FeeRate,
					Creator:   string(address),
					Timestamp: uint64(wh.Timestamp),
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
		if err := c.Call(&result, "eth_getValidator", "0x0"); err != nil {
			return err
		}
		for _, validator := range result.Validators {
			wh.ChangeValidators = append(wh.ChangeValidators, &model.Validator{
				Address: validator.Addr,
				Amount:  validator.Balance.Text(10),
				Proxy:   validator.Proxy,
			})
		}
	}
	return nil
}

// decodeWHTx parses the special transaction of the wormholes blockchain
func decodeWHTx(_ *node.Client, wh *model.Parsed, tx *model.Transaction) (err error) {
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
		Royalty      uint32 `json:"royalty"`
		MetaURL      string `json:"meta_url"`
		FeeRate      uint32 `json:"fee_rate"`
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

	blockNumber := uint64(wh.Number)
	timestamp := uint64(wh.Timestamp)
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

	case 6, 7, 8: //recycle snft, pledge snft, cancel pledge snft
		w.NFTAddress = strings.ToLower(w.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:      w.Type,
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
		if len(w.ProxyAddress) == 42 && w.ProxyAddress != "0x0000000000000000000000000000000000000000" {
			validator.Proxy = w.ProxyAddress
		}
		wh.ChangeValidators = append(wh.ChangeValidators, validator)

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
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		w.Seller2.Exchanger = strings.ToLower(w.Seller2.Exchanger)
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
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
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
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
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		nftAddr := "" //Calculate fill when inserting into database
		wh.NFTs = append(wh.NFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
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

	case 21: // Exchange pledge
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
			Address: from,
			Amount:  value,
		})

	case 22: //Revoke the exchange pledge
		wh.ChangeExchangers = append(wh.ChangeExchangers, &model.Exchanger{
			Address: from,
			Amount:  "-" + value,
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
