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
	"server/service"
)

// DecodeBlock parses the block
func DecodeBlock(c *node.Client, ctx context.Context, number Uint64, isDebug, isWormholes bool) (*service.DecodeRet, error) {
	if n, _ := c.BlockNumber(ctx); n <= uint64(number) {
		return nil, node.NotFound
	}
	var raw json.RawMessage
	// Get the block (including the transaction)
	err := c.CallContext(ctx, &raw, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if len(raw) == 0 {
		return nil, node.NotFound
	}
	block := service.DecodeRet{AddBalance: new(big.Int)}
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	}
	// Get transaction receipt (including transaction log)
	if block.TotalTransaction = Uint64(len(block.CacheTxs)); block.TotalTransaction > 0 {
		// get transaction receipt
		reqs := make([]node.BatchElem, block.TotalTransaction)
		for i, tx := range block.CacheTxs {
			reqs[i] = node.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{tx.Hash},
				Result: &block.CacheTxs[i].Receipt,
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
		err := c.CallContext(ctx, &block.CacheLogs, "eth_getLogs", map[string]interface{}{"blockHash": block.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
	}
	// get uncle block
	if block.UnclesCount = Uint64(len(block.UncleHashes)); block.UnclesCount > 0 {
		block.CacheUncles = make([]*model.Uncle, block.UnclesCount)
		reqs := make([]node.BatchElem, block.UnclesCount)
		for i := range reqs {
			reqs[i] = node.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{block.Hash, Uint64(i)},
				Result: &block.CacheUncles[i],
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
	// Parse changed account properties and internal transactions
	if isDebug {
		err = decodeAccounts(c, ctx, &block)
		if err != nil {
			return nil, err
		}
		for _, tx := range block.CacheTxs {
			internalTxs, err := c.GetInternalTx(ctx, tx)
			if err != nil {
				return nil, fmt.Errorf("GetInternalTx err:%v", err)
			}
			block.CacheInternalTxs = append(block.CacheInternalTxs, internalTxs...)
		}
	}
	// Parse things specific to wormholes
	if isWormholes {
		err = decodeWH(c, &block)
	}
	return &block, err
}

// decodeAccount to get account related properties
func decodeAccounts(c *node.Client, ctx context.Context, block *service.DecodeRet) (err error) {
	block.CacheAccounts = make(map[Address]*model.Account)
	if block.Number > 0 {
		// Get the change account address
		var modifiedAccounts []Address
		err = c.CallContext(ctx, &modifiedAccounts, "debug_getModifiedAccountsByHash", block.Hash)
		if err != nil {
			return fmt.Errorf("debug_getModifiedAccountsByHash err:%v", err)
		}
		for i := range modifiedAccounts {
			// Ignore NFT type addresses
			if modifiedAccounts[i][:14] != "0x000000000000" && modifiedAccounts[i][:14] != "0x800000000000" {
				block.CacheAccounts[modifiedAccounts[i]] = &model.Account{Address: modifiedAccounts[i]}
			}
		}
		for _, tx := range block.CacheTxs {
			if tx.ContractAddress != nil {
				block.CacheAccounts[*tx.ContractAddress] = &model.Account{Address: *tx.ContractAddress, Creator: &tx.From, CreatedTx: &tx.Hash}
			}
		}
		if len(block.CacheAccounts) > 0 {
			// Get account attributes (balance, nonce, code)
			reqs := make([]node.BatchElem, 0, 3*len(block.CacheAccounts))
			for _, account := range block.CacheAccounts {
				reqs = append(reqs, node.BatchElem{
					Method: "eth_getBalance",
					Args:   []interface{}{account.Address, "latest"},
					Result: &account.Balance,
				})
				reqs = append(reqs, node.BatchElem{
					Method: "eth_getTransactionCount",
					Args:   []interface{}{account.Address, "latest"},
					Result: &account.Nonce,
				})
				reqs = append(reqs, node.BatchElem{
					Method: "eth_getCode",
					Args:   []interface{}{account.Address, "latest"},
					Result: &account.Code,
				})
			}
			if err = c.BatchCallContext(ctx, reqs); err != nil {
				return fmt.Errorf("eth_getBalance or eth_getTransactionCount err:%v", err)
			}
			for i := range reqs {
				if reqs[i].Error != nil {
					return fmt.Errorf("eth_getBalance or eth_getTransactionCount err:%v", reqs[i].Error)
				}
			}
		}
	} else {
		genesisAccounts := struct {
			Accounts map[Address]struct {
				Balance BigInt  `json:"balance"`
				Nonce   Uint64  `json:"transactionCount"`
				Code    *string `json:"code"`
			}
		}{}
		err = c.CallContext(ctx, &genesisAccounts, "debug_dumpBlock", block.Number.Hex())
		if err != nil {
			return fmt.Errorf("debug_dumpBlock err:%v", err)
		}
		for address, account := range genesisAccounts.Accounts {
			block.CacheAccounts[address] = &model.Account{
				Address: address,
				Balance: account.Balance,
				Nonce:   account.Nonce,
				Code:    account.Code,
			}
			t, _ := new(big.Int).SetString(string(account.Balance), 10)
			block.AddBalance = block.AddBalance.Add(block.AddBalance, t)
		}
	}
	for address, account := range block.CacheAccounts {
		if account.Code != nil && *account.Code == "0x" {
			account.Code = nil
		}
		account.Name, account.Symbol, err = utils.Property(c, address)
		if err != nil {
			return
		}
		ok, err := utils.IsERC165(c, address)
		if err != nil {
			return err
		}
		if !ok {
			ok, err = utils.IsERC20(c, address)
			if err != nil {
				return err
			}
			if ok {
				account.ERC = ERC20
			}
			continue
		}
		ok, err = utils.IsERC721(c, address)
		if err != nil {
			return err
		}
		if ok {
			account.ERC = ERC721
			continue
		}
		ok, err = utils.IsERC1155(c, address)
		if err != nil {
			return err
		}
		if ok {
			account.ERC = ERC1155
		}
	}
	return nil
}

// decodeWH Imports the underlying NFT transaction of the SNFT meta information distributed by the block
func decodeWH(c *node.Client, wh *service.DecodeRet) error {
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
			case 0, 1, 2, 3:
				identity = 3
			case 4, 5, 6, 7, 8, 9:
				identity = 2
			case 10:
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
					Awardee:      &rewards[i].Address,
					RewardAt:     (*uint64)(&wh.Block.Timestamp),
					RewardNumber: (*uint64)(&wh.Block.Number),
					Owner:        &rewards[i].Address,
				})
				// Parse the new phase ID
				if len(epochId) == 0 {
					addr, _ := new(big.Int).SetString(nftAddr[3:], 16)
					if addr.Mod(addr, big.NewInt(65536)).Uint64() == 0 {
						epochId = nftAddr[:38]
					}
				}
			} else {
				wh.Rewards[i].Amount = new(string)
				*wh.Rewards[i].Amount = rewards[i].RewardAmount.Text(10)
				// Reward an ERB equivalent to SNFT
				wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level0Reward)
			}
		}

		// wormholes transaction processing
		for _, tx := range wh.CacheTxs {
			err = decodeWHTx(c, wh.Block, tx, wh)
			if err != nil {
				return err
			}
		}
	}
	// Write the current information, once every 65536 SNFT rewards
	if len(epochId) > 0 {
		epoch, err := c.GetEpoch((wh.Block.Number - 1).Hex())
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

var (
	level0Reward, _ = new(big.Int).SetString("100000000000000000", 10)       //1
	level1Reward, _ = new(big.Int).SetString("2400000000000000000", 10)      //16
	level2Reward, _ = new(big.Int).SetString("57600000000000000000", 10)     //256
	level3Reward, _ = new(big.Int).SetString("1382400000000000000000", 10)   //4096
	level4Reward, _ = new(big.Int).SetString("221184000000000000000000", 10) //65536
)

// decodeWHTx parses the special transaction of the wormholes blockchain
func decodeWHTx(c *node.Client, block *model.Block, tx *model.Transaction, wh *service.DecodeRet) (err error) {
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
	amount := "0"
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
		level := 42 - len(w.NFTAddress)
		if level == 0 {
			wh.RecycleSNFTs = append(wh.RecycleSNFTs, w.NFTAddress)
		} else {
			// Synthetic SNFT address processing
			for i := 0; i < 1<<(level*4); i++ {
				address := fmt.Sprintf("%s%0"+strconv.Itoa(level)+"x", w.NFTAddress, i)
				wh.RecycleSNFTs = append(wh.RecycleSNFTs, address)
			}
		}
		switch level {
		case 0:
			wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level0Reward)
		case 1:
			wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level1Reward)
		case 2:
			wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level2Reward)
		case 3:
			wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level3Reward)
		case 4:
			wh.AddBalance = wh.AddBalance.Add(wh.AddBalance, level4Reward)
		default:
			return fmt.Errorf("recycle %s SNFT level not support:%d", w.NFTAddress, level)
		}
		return

	case 9, 10: //Consensus pledge, can be pledged multiple times, starting at 100000ERB
		amount, err = c.GetPledge(from, wh.Number.Hex())
		if err != nil {
			return err
		}
		wh.ConsensusPledges = append(wh.ConsensusPledges, &model.ConsensusPledge{
			Address: from,
			Amount:  amount,
			Count:   1,
		})
		internalTx := &model.InternalTx{
			ParentTxHash: Hash(txHash),
			From:         &tx.From,
			To:           tx.To,
			Value:        tx.Value,
			GasLimit:     tx.Gas,
		}
		if w.Type == 9 {
			internalTx.Op = "pledge_add"
		} else {
			internalTx.Op = "pledge_sub"
		}
		wh.CacheInternalTxs = append(wh.CacheInternalTxs, internalTx)

	case 11: //Open the exchange
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.ExchangerPledge{
			Address: from,
			Amount:  value,
		})
		wh.Exchangers = append(wh.Exchangers, &model.Exchanger{
			Address:      from,
			Name:         w.Name,
			URL:          w.Url,
			FeeRatio:     w.FeeRate, //unit 1/10,000
			Creator:      from,
			Timestamp:    timestamp,
			BlockNumber:  blockNumber,
			TxHash:       txHash,
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
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.ExchangerPledge{
			Address: from,
			Amount:  value,
		})

	case 22: //Revoke the exchange pledge
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.ExchangerPledge{
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
