package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"server/common/model"
	. "server/common/types"
	"server/common/utils"
	"server/node"
	"server/service"
)

// DecodeBlock 解析区块
func DecodeBlock(c *node.Client, ctx context.Context, number Uint64) (*service.DecodeRet, error) {
	var raw json.RawMessage
	// 获取区块及其交易
	err := c.CallContext(ctx, &raw, "eth_getBlockByNumber", number.Hex(), true)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	} else if len(raw) == 0 {
		return nil, node.NotFound
	}

	var block service.DecodeRet
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber err:%v", err)
	}

	if totalTransaction := len(block.CacheTxs); totalTransaction > 0 {
		block.TotalTransaction = Uint64(totalTransaction)
		// 获取交易收据
		reqs := make([]node.BatchElem, totalTransaction)
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
		// 获取收据logs,只能根据区块哈希查，区块高度会查到
		err := c.CallContext(ctx, &block.CacheLogs, "eth_getLogs", map[string]interface{}{"blockHash": block.Hash})
		if err != nil {
			return nil, fmt.Errorf("eth_getLogs err:%v", err)
		}
		// 获取解析内部交易
		//for _, tx := range block.CacheTxs {
		//	to := tx.To
		//	if to == nil {
		//		to = tx.ContractAddress
		//	}
		//	internalTxs, err := c.GetInternalTx(ctx, number, tx.Hash, *to)
		//	if err != nil {
		//		return nil, err
		//	}
		//	block.CacheInternalTxs = append(block.CacheInternalTxs, internalTxs...)
		//}
	}

	// 获取叔块
	block.UnclesCount = Uint64(len(block.UncleHashes))
	if block.UnclesCount > 0 {
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

	// 解析相关账户和创建的合约
	block.CacheAccounts = make(map[Address]*model.Account)
	block.CacheContracts = make(map[Address]*model.Contract)
	block.CacheAccounts[block.Miner] = &model.Account{Address: block.Miner}
	if len(block.CacheTxs) > 0 {
		for _, tx := range block.CacheTxs {
			block.CacheAccounts[tx.From] = &model.Account{Address: tx.From}
			if tx.To != nil {
				block.CacheAccounts[*tx.To] = &model.Account{Address: *tx.To}
			}
			if tx.ContractAddress != nil {
				block.CacheAccounts[*tx.ContractAddress] = &model.Account{Address: *tx.ContractAddress}
				block.CacheContracts[*tx.ContractAddress] = &model.Contract{Address: *tx.ContractAddress, Creator: tx.From, CreatedTx: tx.Hash}
			}
		}
	}

	// 获取账户属性
	reqs := make([]node.BatchElem, 0, 2*len(block.CacheAccounts))
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
	}
	if err := c.BatchCallContext(ctx, reqs); err != nil {
		return nil, fmt.Errorf("eth_getBalance or eth_getTransactionCount err:%v", err)
	}
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, fmt.Errorf("eth_getBalance or eth_getTransactionCount err:%v", reqs[i].Error)
		}
	}

	// 获取合约属性
	if len(block.CacheContracts) > 0 {
		reqs := make([]node.BatchElem, 0, len(block.CacheContracts))
		for _, contract := range block.CacheContracts {
			reqs = append(reqs, node.BatchElem{
				Method: "eth_getCode",
				Args:   []interface{}{contract.Address, "latest"},
				Result: &contract.Code,
			})
		}
		if err := c.BatchCallContext(ctx, reqs); err != nil {
			return nil, fmt.Errorf("eth_getCode err:%v", err)
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, fmt.Errorf("eth_getCode err:%v", reqs[i].Error)
			}
		}

		for _, contract := range block.CacheContracts {
			code, _ := hex.DecodeString(contract.Code[2:])
			hash := utils.Keccak256Hash(code)
			block.CacheAccounts[contract.Address].CodeHash = &hash
			if len(contract.Code) > 2 {
				contract.ERC, err = utils.GetERC(c, contract.Address)
				if err != nil {
					return nil, fmt.Errorf("GetERC() err:%v", err)
				}
			}
		}
	}

	err = decodeWH(c, &block)
	return &block, err
}

// decodeWH 导入区块分发的SNFT元信息底层NFT交易
func decodeWH(c *node.Client, wh *service.DecodeRet) (err error) {
	if wh.Number == 0 {
		return
	}
	// 矿工奖励SNFT处理
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
			err = fmt.Errorf("reward length more than 11")
			return
		}
		wh.Rewards = append(wh.Rewards, &model.Reward{
			Address:     rewards[i].Address,
			Identity:    identity,
			BlockNumber: uint64(wh.Block.Number),
			SNFT:        rewards[i].NFTAddress,
			Amount:      rewards[i].RewardAmount,
		})
		if rewards[i].NFTAddress != nil {
			wh.RewardSNFTs = append(wh.RewardSNFTs, &model.SNFT{
				Address:      *rewards[i].NFTAddress,
				Awardee:      &rewards[i].Address,
				RewardAt:     (*uint64)(&wh.Block.Timestamp),
				RewardNumber: (*uint64)(&wh.Block.Number),
				Owner:        &rewards[i].Address,
			})
		}
	}
	//解决NFT元信息等没有注入问题(包含创世区块的)，正常应该解析官方注入InjectSNFT的交易
	if len(wh.RewardSNFTs) > 0 {
		var lastAddr = wh.RewardSNFTs[len(wh.RewardSNFTs)-1].Address
		snft, err := c.GetSNFT(lastAddr, wh.Block.Number.Hex())
		if err != nil {
			return fmt.Errorf("GetSNFT() err:%v", err)
		}

		epoch := "0x8" + lastAddr[3:38]
		dir := ""
		if len(snft.MetaURL) > 53 {
			dir = snft.MetaURL[0:53]
		}
		wh.Epochs = append(wh.Epochs, &model.Epoch{ID: epoch, RoyaltyRatio: snft.Royalty, Dir: dir, Creator: snft.Creator})
	}
	// wormholes交易处理
	for _, tx := range wh.CacheTxs {
		err = decodeWHTx(wh.Block, tx, wh)
		if err != nil {
			return
		}
	}
	return
}

// decodeWHTx 解析wormholes区块链的特殊交易
func decodeWHTx(block *model.Block, tx *model.Transaction, wh *service.DecodeRet) (err error) {
	input, _ := hex.DecodeString(tx.Input[2:])
	// 非wormholes类型和失败的交易不解析
	if len(input) < 10 || string(input[0:10]) != "wormholes:" || *tx.Status == 0 {
		return
	}
	type Wormholes struct {
		Type       uint8  `json:"type"`
		NFTAddress string `json:"nft_address"`
		Exchanger  string `json:"exchanger"`
		Royalty    uint32 `json:"royalty"`
		MetaURL    string `json:"meta_url"`
		FeeRate    uint32 `json:"fee_rate"`
		Name       string `json:"name"`
		Url        string `json:"url"`
		Dir        string `json:"dir"`
		StartIndex string `json:"start_index"`
		Number     uint64 `json:"number"`
		Buyer      struct {
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
		Creator string `json:"creator"`
		Version string `json:"version"`
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
	case 0: //用户自行铸造NFT
		nftAddr := "" //插入数据库时实时计算填充
		wh.CreateNFTs = append(wh.CreateNFTs, &model.UserNFT{
			Address:       &nftAddr,
			RoyaltyRatio:  w.Royalty, //单位万分之一
			MetaUrl:       realMeatUrl(w.MetaURL),
			RawMetaUrl:    w.MetaURL,
			ExchangerAddr: strings.ToLower(w.Exchanger),
			Creator:       to,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         to,
		})

	case 1: //NFT自行转移
		w.NFTAddress = strings.ToLower(w.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        1,
			NFTAddr:       &w.NFTAddress,
			ExchangerAddr: "", //自行转移没有交易所
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         nil,
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 6: //官方NFT兑换,回收碎片到碎片池
		wh.RecycleSNFTs = append(wh.RecycleSNFTs, w.NFTAddress)
		return

	case 9: //共识质押
		wh.ConsensusPledges = append(wh.ConsensusPledges, &model.ConsensusPledge{
			Address: from,
			Amount:  value,
		})

	case 10: //撤销共识质押
		wh.ConsensusPledges = append(wh.ConsensusPledges, &model.ConsensusPledge{
			Address: from,
			Amount:  "-" + value,
		})

	case 11: //开启交易所
		wh.Exchangers = append(wh.Exchangers, &model.Exchanger{
			Address:     from,
			Name:        w.Name,
			URL:         w.Url,
			FeeRatio:    w.FeeRate, //单位万分之一
			Creator:     from,
			Timestamp:   timestamp,
			IsOpen:      true,
			BlockNumber: blockNumber,
			TxHash:      txHash,
		})

	case 12: //关闭交易所
		wh.Exchangers = append(wh.Exchangers, &model.Exchanger{
			Address: from,
			IsOpen:  false,
		})

	case 13: //官方注入NFT
		epoch := "0x8" + w.StartIndex[3:38]
		log.Println("官方注入:", w.StartIndex, w.Number, w.Royalty, w.Creator, w.Dir)
		wh.Epochs = append(wh.Epochs, &model.Epoch{ID: epoch, RoyaltyRatio: w.Royalty, Dir: w.Dir, Creator: w.Creator, Number: blockNumber, Timestamp: timestamp, TxHash: txHash})

	case 14: //NFT出价成交交易（卖家或交易所发起,买家给价格签名）
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		w.Buyer.Exchanger = strings.ToLower(w.Buyer.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        2,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: w.Buyer.Exchanger,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 15: //NFT定价购买交易（买家发起，卖家给价格签名）
		w.Seller1.NFTAddress = strings.ToLower(w.Seller1.NFTAddress)
		w.Seller1.Exchanger = strings.ToLower(w.Seller1.Exchanger)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        3,
			NFTAddr:       &w.Seller1.NFTAddress,
			ExchangerAddr: w.Seller1.Exchanger,
			From:          "",     //插入数据库时实时填充原拥有者
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 16: //NFT惰性定价购买交易，买家发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		w.Seller2.Exchanger = strings.ToLower(w.Seller2.Exchanger)
		nftAddr := "" //插入数据库时计算填充
		wh.CreateNFTs = append(wh.CreateNFTs, &model.UserNFT{
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
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 17: //NFT惰性定价购买交易，交易所发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		nftAddr := "" //插入数据库时计算填充
		wh.CreateNFTs = append(wh.CreateNFTs, &model.UserNFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: from, //交易发起者即交易所地址
			Creator:       string(creator),
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         string(creator),
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        5,
			NFTAddr:       &nftAddr,
			ExchangerAddr: from, //交易发起者即交易所地址
			From:          string(creator),
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 18: //NFT出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从授权签名恢复交易所地址
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
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 19: //NFT惰性出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := utils.RecoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// 从授权签名恢复交易所地址
		msg = w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := utils.RecoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return err
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return err
		}
		nftAddr := "" //插入数据库时计算填充
		wh.CreateNFTs = append(wh.CreateNFTs, &model.UserNFT{
			Address:       &nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			RawMetaUrl:    w.Seller2.MetaURL,
			ExchangerAddr: string(exchangerAddr), //交易发起者即交易所地址
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
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 20: //NFT撮合交易，交易所发起
		w.Buyer.NFTAddress = strings.ToLower(w.Buyer.NFTAddress)
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        8,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: from,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
			BlockNumber:   blockNumber,
		})

	case 21: //交易所质押
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.ExchangerPledge{
			Address: from,
			Amount:  value,
		})

	case 22: //撤销交易所质押
		wh.ExchangerPledges = append(wh.ExchangerPledges, &model.ExchangerPledge{
			Address: from,
			Amount:  "-" + value,
		})
	}
	return
}

// realMeatUrl 解析真正的metaUrl
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
