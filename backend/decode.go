package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"server/common/model"
	. "server/common/types"
	"server/common/utils"
	"server/node"
)

// decode parses the block
func decode(c *node.Client, ctx context.Context, number Uint64, isDebug, isWormholes bool) (*model.Parsed, error) {
	var raw json.RawMessage
	// Get the block (including the transaction)
	err := c.CallContext(ctx, &raw, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if len(raw) == 0 {
		return nil, node.NotFound
	}
	var parsed model.Parsed
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	}
	// Get transaction receipt (including transaction log)
	if parsed.TotalTransaction = Uint64(len(parsed.CacheTxs)); parsed.TotalTransaction > 0 {
		// get transaction receipt
		reqs := make([]node.BatchElem, parsed.TotalTransaction)
		for i, tx := range parsed.CacheTxs {
			reqs[i] = node.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{tx.Hash},
				Result: &parsed.CacheTxs[i].Receipt,
			}
		}
		if err := c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getTransactionReceipt err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, fmt.Errorf("eth_getTransactionReceipt err:%v", reqs[i].Error)
			}
		}
		// Get the receipt logs, which can only be checked according to the block hash (there may be multiple blocks with the same block height)
		err := c.CallContext(ctx, &parsed.CacheLogs, "eth_getLogs", map[string]interface{}{"blockHash": parsed.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
	}
	// get uncle block
	if parsed.UnclesCount = Uint64(len(parsed.UncleHashes)); parsed.UnclesCount > 0 {
		parsed.CacheUncles = make([]*model.Uncle, parsed.UnclesCount)
		reqs := make([]node.BatchElem, parsed.UnclesCount)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{parsed.Hash, Uint64(i)},
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
	if isDebug {
		err = decodeAccounts(c, ctx, &parsed)
		if err != nil {
			return nil, fmt.Errorf("decodeAccounts err:%v", err)
		}
		err = decodeInternalTxs(c, ctx, &parsed)
		if err != nil {
			return nil, fmt.Errorf("decodeInternalTxs err:%v", err)
		}
	}
	// Parse things specific to wormholes
	if isWormholes {
		err = decodeWH(c, &parsed)
	}
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
		callers := []*Address{caller}
		iTx := &model.InternalTx{TxHash: tx.Hash, BlockNumber: parsed.Number, To: new(Address), Value: "0x0"}
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
			iTx.Depth = Uint64(log.Depth)
			iTx.Op = log.Op
			iTx.From = callers[log.Depth-1]
			iTx.GasLimit = Uint64(log.Gas)
			parsed.CacheInternalTxs = append(parsed.CacheInternalTxs, iTx)
			iTx = &model.InternalTx{TxHash: tx.Hash, BlockNumber: parsed.Number, To: new(Address), Value: "0x0"}
		}
	}
	return nil
}

// decodeAccount to get account related properties
func decodeAccounts(c *node.Client, ctx context.Context, parsed *model.Parsed) (err error) {
	parsed.CacheAccounts = make(map[Address]*model.Account)
	if parsed.Number > 0 {
		// Get the change account address
		var modifiedAccounts []Address
		if err = c.CallContext(ctx, &modifiedAccounts, "debug_getModifiedAccountsByHash", parsed.Hash); err != nil {
			return
		}
		for _, address := range modifiedAccounts {
			if address[:14] != "0x000000000000" && address[:14] != "0x800000000000" {
				parsed.CacheAccounts[address] = &model.Account{Address: address}
			}
		}
	} else {
		genesis := &struct {
			Accounts map[Address]struct{} `json:"accounts"`
			Next     *string              `json:"next"`
		}{}
		for next := new(string); next != nil; next = genesis.Next {
			genesis = nil
			if err = c.CallContext(ctx, &genesis, "debug_accountRange", "0x0", next, nil, true, true, true); err != nil {
				return
			}
			for address := range genesis.Accounts {
				parsed.CacheAccounts[address] = &model.Account{Address: address}
			}
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

		var contracts []Address
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

// decodeWH Imports the underlying NFT transaction of the SNFT meta information distributed by the block
func decodeWH(c *node.Client, wh *model.Parsed) error {
	epochId := ""
	if wh.Number > 0 {
		// Miner reward SNFT processing
		rewards, err := c.GetReward(wh.Block.Number.Hex())
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
				Proxy:       rewards[i].Proxy,
				Identity:    identity,
				BlockNumber: uint64(wh.Block.Number),
			})
			if rewards[i].RewardAmount == nil {
				// Note that when NFTAddress is zero address error
				wh.Rewards[i].SNFT = rewards[i].NFTAddress
				nftAddr := *rewards[i].NFTAddress
				wh.RewardSNFTs = append(wh.RewardSNFTs, &model.SNFT{
					Address:      nftAddr,
					Awardee:      rewards[i].Address,
					RewardAt:     uint64(wh.Block.Timestamp),
					RewardNumber: uint64(wh.Block.Number),
					Owner:        rewards[i].Address,
				})
				// Parse the new phase ID
				if len(epochId) == 0 {
					addr, _ := new(big.Int).SetString(nftAddr[3:], 16)
					if addr.Mod(addr, big.NewInt(4096)).Uint64() == 0 {
						epochId = nftAddr[:39]
					}
				}
			} else {
				wh.Rewards[i].Amount = new(string)
				*wh.Rewards[i].Amount = rewards[i].RewardAmount.Text(10)
			}
		}

		// wormholes transaction processing
		for _, tx := range wh.CacheTxs {
			err = decodeWHTx(c, wh.Block, tx, wh)
			if err != nil {
				return err
			}
		}
	} else {
		validators := struct {
			Validators []*struct {
				Addr    string   `json:"Addr"`
				Balance *big.Int `json:"Balance"`
				Proxy   string   `json:"Proxy"`
			} `json:"Validators"`
		}{}
		err := c.Call(&validators, "eth_getValidator", "0x0")
		if err != nil {
			return err
		}
		for _, validator := range validators.Validators {
			wh.ConsensusPledges = append(wh.ConsensusPledges, &model.Pledge{
				Address: validator.Addr,
				Amount:  validator.Balance.Text(10),
			})
		}
	}
	// Write the current information, once every 4096 SNFT rewards
	if len(epochId) > 0 {
		epoch, err := c.GetEpoch(wh.Block.Number.Hex())
		if err != nil {
			return fmt.Errorf("GetEpoch() err:%v", err)
		}
		if len(epoch.Dir) == 52 {
			epoch.Dir = epoch.Dir + "/"
		}
		wh.Epochs = append(wh.Epochs, &model.Epoch{
			ID:           epochId,
			Creator:      strings.ToLower(epoch.Creator),
			RoyaltyRatio: epoch.Royalty,
			Dir:          epoch.Dir,
			Exchanger:    epoch.Address,
			VoteWeight:   epoch.VoteWeight.Text(10),
			Number:       uint64(wh.Block.Number),
			Timestamp:    uint64(wh.Block.Timestamp),
		})
	}
	return nil
}

// decodeWHTx parses the special transaction of the wormholes blockchain
func decodeWHTx(_ *node.Client, block *model.Block, tx *model.Transaction, wh *model.Parsed) (err error) {
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

	blockNumber := uint64(block.Number)
	timestamp := uint64(block.Timestamp)
	txHash := string(tx.Hash)
	from := string(tx.From)
	to := string(*tx.To)
	value := string(tx.Value)
	switch w.Type {
	case 0: //Users mint NFT by themselves
		nftAddr := "" //Calculate fill in real time when inserting into database
		wh.CreateNFTs = append(wh.CreateNFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  w.Royalty, //The unit is one ten thousandth
			MetaUrl:       realMeatUrl(w.MetaURL),
			RawMetaUrl:    w.MetaURL,
			ExchangerAddr: strings.ToLower(w.Exchanger),
			Creator:       to,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         to,
		})

	case 1: //Users transfer NFT by themselves
		w.NFTAddress = strings.ToLower(w.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        1,
			NFTAddr:       &w.NFTAddress,
			ExchangerAddr: "", //Self-transfer without exchange
			From:          "", //The original owner is populated in real time when inserting into the database
			To:            to,
			Price:         nil,
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 6: //Official NFT exchange, recycling shards to shard pool
		wh.RecycleSNFTs = append(wh.RecycleSNFTs, w.NFTAddress)

	case 7: //pledge snft
		level := 42 - len(w.NFTAddress)
		if level == 0 {
			wh.PledgeSNFT = append(wh.PledgeSNFT, w.NFTAddress)
		} else {
			for i := 0; i < 1<<(level*4); i++ {
				address := fmt.Sprintf("%s%0"+strconv.Itoa(level)+"x", w.NFTAddress, i)
				wh.PledgeSNFT = append(wh.PledgeSNFT, address)
			}
		}

	case 8: //cancel pledge snft
		level := 42 - len(w.NFTAddress)
		if level == 0 {
			wh.UnPledgeSNFT = append(wh.UnPledgeSNFT, w.NFTAddress)
		} else {
			for i := 0; i < 1<<(level*4); i++ {
				address := fmt.Sprintf("%s%0"+strconv.Itoa(level)+"x", w.NFTAddress, i)
				wh.UnPledgeSNFT = append(wh.UnPledgeSNFT, address)
			}
		}

	case 9, 10: //Consensus pledge, can be pledged multiple times, starting at 100000ERB
		pledge := &model.Pledge{
			Address: from,
		}
		internalTx := &model.InternalTx{
			TxHash:      Hash(txHash),
			BlockNumber: wh.Number,
			From:        &tx.From,
			To:          tx.To,
			Value:       tx.Value,
			GasLimit:    tx.Gas,
		}
		if w.Type == 9 {
			internalTx.Op = "pledge_add"
			pledge.Amount = value
		} else {
			internalTx.Op = "pledge_sub"
			pledge.Amount = "-" + value

		}
		wh.ConsensusPledges = append(wh.ConsensusPledges, pledge)
		wh.CacheInternalTxs = append(wh.CacheInternalTxs, internalTx)

	case 11: //Open the exchange
		wh.Exchangers = append(wh.Exchangers, &model.Exchanger{
			Address:      from,
			Name:         w.Name,
			URL:          w.Url,
			FeeRatio:     w.FeeRate, //unit 1/10,000
			Creator:      from,
			Timestamp:    timestamp,
			BlockNumber:  blockNumber,
			TxHash:       txHash,
			Amount:       value,
			Count:        1,
			BalanceCount: "0",
		})

	case 12: //Close the exchange
		wh.CloseExchangers = append(wh.CloseExchangers, from)

	case 14: //NFT bid transaction (initiated by the seller or the exchange, and the buyer signs the price)
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		w.Buyer.Exchanger = strings.ToLower(w.Buyer.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        2,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: w.Buyer.Exchanger,
			From:          "", //The original owner is populated in real time when inserting into the database
			To:            to,
			Price:         &value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 15: //NFT pricing purchase transaction (buyer initiates, seller signs price)
		w.Seller1.NFTAddress = strings.ToLower(w.Seller1.NFTAddress)
		w.Seller1.Exchanger = strings.ToLower(w.Seller1.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        3,
			NFTAddr:       &w.Seller1.NFTAddress,
			ExchangerAddr: w.Seller1.Exchanger,
			From:          "",     //The original owner is populated in real time when inserting into the database
			To:            from,   //The transaction initiator is the buyer
			Price:         &value, //The unit is wei
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
		wh.CreateNFTs = append(wh.CreateNFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: w.Seller2.Exchanger,
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        4,
			NFTAddr:       &nftAddr,
			ExchangerAddr: w.Seller2.Exchanger,
			From:          string(creator),
			To:            from,   //The transaction initiator is the buyer
			Price:         &value, //The unit is wei
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
		wh.CreateNFTs = append(wh.CreateNFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: from, //The transaction initiator is the exchange address
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        5,
			NFTAddr:       &nftAddr,
			ExchangerAddr: from, //The transaction initiator is the exchange address
			From:          string(creator),
			To:            to,
			Price:         &value, //The unit is wei
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
			TxType:        6,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: string(exchangerAddr),
			From:          "", //The original owner is populated in real time when inserting into the database
			To:            to,
			Price:         &value, //The unit is wei
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
		wh.CreateNFTs = append(wh.CreateNFTs, &model.NFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: string(exchangerAddr), //The transaction initiator is the exchange address
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        7,
			NFTAddr:       &nftAddr,
			ExchangerAddr: string(exchangerAddr),
			From:          string(creator),
			To:            to,
			Price:         &value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 20: //NFT matches the transaction, and the exchange initiates it
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        8,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: from,
			From:          "", //The original owner is populated in real time when inserting into the database
			To:            to,
			Price:         &value, //The unit is wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 21: // Exchange pledge
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.Exchanger{
			Address: from,
			Amount:  value,
		})

	case 22: //Revoke the exchange pledge
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.Exchanger{
			Address: from,
			Amount:  "-" + value,
		})
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
