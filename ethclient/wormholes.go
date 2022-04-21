package ethclient

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"server/model"
)

type WH struct {
	CreateSNFTs      []*model.OfficialNFT     //SNFT的创建信息
	RewardSNFTs      []*model.OfficialNFT     //SNFT的奖励信息
	CreateNFTs       []*model.UserNFT         //新创建的NFT
	NFTTxs           []*model.NFTTx           //NFT交易记录
	Exchanger        []*model.Exchanger       //交易所
	RecycleSNFTs     []string                 //回收的SNFT
	InjectSNFTs      []*InjectSNFT            //官方注入SNFT
	ExchangerPledges []*model.ExchangerPledge //交易所质押
	ConsensusPledges []*model.ConsensusPledge //共识质押
}

type InjectSNFT struct {
	StartIndex        *big.Int
	Count             uint64
	Royalty           uint32
	Dir, Creator      string
	Number, Timestamp uint64
}

// DecodeWH 解析wormholes链独有的东西
func (ec *Client) DecodeWH(block *model.Block) (wh WH, err error) {
	// 创世区块到入预设SNFT
	if block.Number == 0 {
		wh.CreateSNFTs, err = ec.DecodeGenesisSNFT(block)
		log.Println(wh.CreateSNFTs, err)
	} else {
		wh.RewardSNFTs, wh.CreateSNFTs, err = ec.DecodeBlockSNFT(block)
	}
	return
}

// DecodeGenesisSNFT 导入创世区块注入的SNFT元信息
func (ec *Client) DecodeGenesisSNFT(block *model.Block) (snfts []*model.OfficialNFT, err error) {
	SNFTAddr := big.NewInt(0)
	SNFTAddr.SetString("8000000000000000000000000000000000000000", 16)
lo:
	addr := strings.ToLower(common.BigToAddress(SNFTAddr).String())
	snft, err := ec.GetSNFT(addr, "0x0")
	if err != nil {
		return
	}
	if snft.MetaURL != "" {
		snfts = append(snfts, &model.OfficialNFT{
			Address:      addr,
			CreateAt:     uint64(block.Timestamp),
			CreateNumber: 0,
			Creator:      snft.Creator,
			RoyaltyRatio: snft.Royalty,
			MetaUrl:      snft.MetaURL,
		})
		SNFTAddr = SNFTAddr.Add(SNFTAddr, common.Big1)
		goto lo
	}
	return
}

// DecodeBlockSNFT 导入区块分发的SNFT元信息
func (ec *Client) DecodeBlockSNFT(block *model.Block) (rewardSNFTs, createSNFTs []*model.OfficialNFT, err error) {
	rewards, err := ec.GetReward(block.Number.String())
	if err != nil {
		return
	}
	for i := range rewards {
		rewardSNFTs = append(rewardSNFTs, &model.OfficialNFT{
			Address:      rewards[i].NfTAddress,
			Awardee:      &rewards[i].Address,
			RewardAt:     (*uint64)(&block.Timestamp),
			RewardNumber: (*uint64)(&block.Number),
			Owner:        &rewards[i].Address,
		})
		//---todo 临时解决NFT元信息等没有注入问题，正常应该解析官方注入InjectSNFT的交易来填写SNFT元信息----
		var snft SNFT
		snft, err = ec.GetSNFT(rewards[i].NfTAddress, block.Number.String())
		if err != nil {
			return
		}
		createSNFTs = append(createSNFTs, &model.OfficialNFT{
			Address:      rewards[i].NfTAddress,
			CreateAt:     uint64(block.Timestamp),
			CreateNumber: uint64(block.Number),
			Creator:      snft.Creator,
			RoyaltyRatio: snft.Royalty,
			MetaUrl:      snft.MetaURL,
		})
	}
	return
}

// DecodeWHTx 解析wormholes区块链的特殊交易
func (ec *Client) DecodeWHTx(block *model.Block, tx *model.Transaction, wh *WH) (err error) {
	input, _ := hexutil.Decode(tx.Input)
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
	value := tx.Value
	switch w.Type {
	case 0: //用户自行铸造NFT
		nftAddr := "" //插入数据库时实时计算填充
		wh.CreateNFTs = append(wh.CreateNFTs, &model.UserNFT{
			Address:       &nftAddr,
			RoyaltyRatio:  w.Royalty, //单位万分之一
			MetaUrl:       realMeatUrl(w.MetaURL),
			ExchangerAddr: w.Exchanger,
			Creator:       to,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         to,
		})

	case 1: //NFT自行转移
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        1,
			NFTAddr:       &w.NFTAddress,
			ExchangerAddr: nil, //自行转移没有交易所
			From:          "",  //插入数据库时实时填充原拥有者
			To:            to,
			Price:         nil,
			Timestamp:     timestamp,
			TxHash:        txHash,
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
		wh.Exchanger = append(wh.Exchanger, &model.Exchanger{
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
		wh.Exchanger = append(wh.Exchanger, &model.Exchanger{
			Address: from,
			IsOpen:  false,
		})

	case 13: //官方注入NFT
		startIndex, _ := new(big.Int).SetString(w.StartIndex[2:], 16)
		if err != nil {
			return
		}
		fmt.Println("官方注入:", startIndex, w.Number, w.Royalty, w.Creator, w.Dir)
		wh.InjectSNFTs = append(wh.InjectSNFTs, &InjectSNFT{startIndex, w.Number, w.Royalty, w.Dir, w.Creator, blockNumber, timestamp})

	case 14: //NFT出价成交交易（卖家或交易所发起,买家给价格签名）
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        2,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &w.Buyer.Exchanger,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 15: //NFT定价购买交易（买家发起，卖家给价格签名）
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        3,
			NFTAddr:       &w.Seller1.NFTAddress,
			ExchangerAddr: &w.Seller1.Exchanger,
			From:          "",     //插入数据库时实时填充原拥有者
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 16: //NFT惰性定价购买交易，买家发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
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
			ExchangerAddr: w.Seller2.Exchanger,
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        4,
			NFTAddr:       &nftAddr,
			ExchangerAddr: &w.Seller2.Exchanger,
			From:          creator,
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 17: //NFT惰性定价购买交易，交易所发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
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
			ExchangerAddr: from, //交易发起者即交易所地址
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        5,
			NFTAddr:       &nftAddr,
			ExchangerAddr: &from, //交易发起者即交易所地址
			From:          creator,
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 18: //NFT出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从授权签名恢复交易所地址
		msg := w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := recoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return err
		}
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        6,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &exchangerAddr,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 19: //NFT惰性出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return err
		}
		// 从授权签名恢复交易所地址
		msg = w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := recoverAddress(msg, w.ExchangerAuth.Sig)
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
			ExchangerAddr: exchangerAddr, //交易发起者即交易所地址
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		})
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        7,
			NFTAddr:       &nftAddr,
			ExchangerAddr: &exchangerAddr,
			From:          creator,
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		})

	case 20: //NFT撮合交易，交易所发起
		wh.NFTTxs = append(wh.NFTTxs, &model.NFTTx{
			TxType:        8,
			NFTAddr:       &w.Buyer.NFTAddress,
			ExchangerAddr: &from,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
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
