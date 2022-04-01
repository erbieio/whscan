package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"server/database"
	"server/ethhelper"
	"server/log"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

func SyncBlock() {
	number, err := database.BlockCount()
	if err != nil {
		panic(err)
	}
	fmt.Printf("从%v区块开始数据分析\n", number)
	for {
		currentNumber, _ := ethhelper.GetBlockNumber()
		err := HandleBlock(number)
		if err != nil || currentNumber < number {
			fmt.Printf("最新区块高度%v：在%v区块休眠, 错误：%v\n", currentNumber, number, err)
			time.Sleep(5 * time.Second)
		}
		if err == nil {
			number++
		}
	}
}

func HandleBlock(number uint64) error {
	var tmp big.Int
	tmp.SetUint64(number)

	b, err := ethhelper.GetBlock("0x" + tmp.Text(16))
	if err != nil {
		log.Infof(err.Error())
		return err
	}
	var block database.Block
	hexTo10 := func(v string) string {
		tmp.SetString(v, 0)
		return fmt.Sprintf("%v", tmp.Uint64())
	}
	block.ParentHash = b.UncleHash
	block.UncleHash = b.UncleHash
	block.Coinbase = b.Coinbase
	block.Root = b.Root
	block.TxHash = b.TxHash
	block.ReceiptHash = b.ReceiptHash
	block.Difficulty = b.Difficulty
	block.Number = hexTo10(b.Number)
	block.GasLimit = hexTo10(b.GasLimit)
	block.GasUsed = hexTo10(b.GasUsed)
	block.Extra = b.Extra
	block.MixDigest = b.MixDigest
	block.Size = hexTo10(b.Size)
	block.TxCount = len(b.Transactions)
	block.Ts = hexTo10(b.Ts)
	err = block.Insert()
	if err != nil {
		log.Infof(err.Error())
	}
	type ethTransfer struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Value string `json:"value"`
	}
	var eths []ethTransfer
	for _, t := range b.Transactions {
		var tx database.Transaction
		tx.Hash = t.Hash
		tx.From = t.From
		tx.To = t.To
		tx.Value = hexTo10(t.Value)
		tx.Value = toDecimal(tx.Value, 18)
		tx.Nonce = hexTo10(t.Nonce)
		tx.BlockHash = t.BlockHash
		tx.BlockNumber = hexTo10(t.BlockNumber)
		tx.TransactionIndex = hexTo10(t.TransactionIndex)
		tx.Gas = hexTo10(t.Gas)
		tx.GasPrice = hexTo10(t.GasPrice)
		tx.Input = t.Input

		calls := TraceTxInternalCall(tx.Hash, tx.From, tx.To)
		callsStr, _ := json.Marshal(calls)
		tx.InternalCalls = string(callsStr)
		for _, c := range calls {
			var f float64
			f, _ = strconv.ParseFloat(c.Value, 64)
			if f != 0 {
				eths = append(eths, ethTransfer{
					From:  c.From,
					To:    c.To,
					Value: c.Value,
				})
			}
		}
		ethsStr, _ := json.Marshal(eths)
		tx.InternalValueTransfer = string(ethsStr)
		tx.Ts = block.Ts
		status, ty, tokenTransfers := analysisTokenTransfer(tx.Hash)
		if status == "0x1" {
			tx.Status = "1"
		} else {
			tx.Status = "0"
		}

		input, _ := hexutil.Decode(t.Input)
		if len(input) >= 10 && string(input[0:10]) == "wormholes:" {
			// 从input字段判断是否是wormholes交易，并解析处理
			HandleWHTx(input[10:], t.BlockNumber, b.Ts, t.Hash, t.From, t.To, t.Value, status == "0x1")
		}

		tx.TxType = ty
		tokenStr, _ := json.Marshal(tokenTransfers)
		tx.TokenTransfer = string(tokenStr)
		err = tx.Insert()
		if err != nil {
			log.Infof(err.Error())
		}
	}

	handleWHBlock(b.Number, b.Ts)
	return nil
}

// 导入创世区块注入的SNFT元信息
func handleGenesisSNFT(timestamp uint64) error {
	SNFTAddr := big.NewInt(0)
	SNFTAddr.SetString("8000000000000000000000000000000000000000", 16)
lo:
	addr := strings.ToLower(common.BigToAddress(SNFTAddr).String())
	snft, err := ethhelper.GetGenesisSNFT(addr)
	if err != nil {
		return err
	}
	if snft.MetaURL != "" {
		err = database.ImportSNFT(addr, snft.Royalty, "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00", snft.Creator, 0, timestamp)
		// 后台解析meta
		go SaveNFTMeta(0, addr, "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00")
		if err != nil {
			return err
		}
		SNFTAddr = SNFTAddr.Add(SNFTAddr, common.Big1)
		goto lo
	}
	fmt.Println("创世区块SNFT导入到地址：", addr)
	return nil
}

func handleWHBlock(number, time string) {
	var err error
	defer func() {
		if err != nil {
			log.Infof("解析validators", "区块", number, "错误", err)
		}
	}()
	timestamp, err := strconv.ParseUint(time[2:], 16, 64)
	if err != nil {
		return
	}
	blockNumber, err := strconv.ParseUint(number[2:], 16, 64)
	if err != nil {
		return
	}
	// 创世区块到入预设SNFT
	if blockNumber == 0 {
		handleGenesisSNFT(timestamp)
		return
	}
	rewards, err := ethhelper.GetReward(number)
	if err != nil {
		return
	}
	for _, rewards := range rewards {
		err = database.DispatchSNFT(rewards.Address, rewards.NfTAddress, blockNumber, timestamp)
		if err != nil {
			return
		}

		//---临时解决NFT元信息等没有注入问题，正常应该解析官方注入InjectSNFT的交易来填写SNFT元信息----
		snft, err := ethhelper.GetSNFT(rewards.NfTAddress, number)
		if err != nil {
			return
		}
		err = database.SaveSNFT(rewards.NfTAddress, snft.Royalty, "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00", snft.Creator, blockNumber, timestamp)
		if err != nil {
			return
		}
		// 后台解析meta
		go SaveNFTMeta(blockNumber, rewards.NfTAddress, "/ipfs/QmeCPcX3rYguWqJYDmJ6D4qTQqd5asr8gYpwRcgw44WsS7/00")
		//---临时解决-----------------------------------------------------------------------
	}
}

// HandleWHTx 解析wormholes区块链的特殊交易
func HandleWHTx(input []byte, number, time, txHash, from, to, value string, status bool) {
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
	var nft *database.UserNFT
	var err error
	defer func() {
		if err != nil {
			log.Infof("解析wormholes", "区块", number, "交易", txHash, "input", string(input), "错误", err)
		}
	}()
	if err := json.Unmarshal(input, &w); err != nil {
		return
	}

	if status == false {
		err = errors.New("失败的交易不解析")
		return
	}

	timestamp, err := strconv.ParseUint(time[2:], 16, 64)
	if err != nil {
		return
	}
	blockNumber, err := strconv.ParseUint(number[2:], 16, 64)
	if err != nil {
		return
	}

	switch w.Type {
	case 0: //用户自行铸造NFT
		nftAddr, err := database.GetNFTAddr()
		if err != nil {
			return
		}
		nft = &database.UserNFT{
			Address:       nftAddr,
			RoyaltyRatio:  w.Royalty, //单位万分之一
			MetaUrl:       realMeatUrl(w.MetaURL),
			ExchangerAddr: w.Exchanger,
			Creator:       to,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         to,
		}
		err = nft.Insert()
		if err != nil {
			return
		}

	case 1: //NFT自行转移
		nftTx := database.NFTTx{
			TxType:        1,
			NFTAddr:       w.NFTAddress,
			ExchangerAddr: nil, //自行转移没有交易所
			From:          "",  //插入数据库时实时填充原拥有者
			To:            to,
			Price:         nil,
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 6: //官方NFT兑换,回收碎片到碎片池
		database.RecycleSNFT(w.NFTAddress)
		return

	case 9: //共识质押
		err = database.ConsensusPledgeAdd(from, value)
		if err != nil {
			return
		}

	case 10: //撤销共识质押
		err = database.ConsensusPledgeSub(from, value)
		if err != nil {
			return
		}

	case 11: //开启交易所
		exchanger := database.Exchanger{
			Address:     from,
			Name:        w.Name,
			URL:         w.Url,
			FeeRatio:    w.FeeRate, //单位万分之一
			Creator:     from,
			Timestamp:   timestamp,
			IsOpen:      true,
			BlockNumber: blockNumber,
			TxHash:      txHash,
		}
		err = database.OpenExchange(&exchanger)
		if err != nil {
			return
		}

	case 12: //关闭交易所
		err = database.CloseExchange(from)
		if err != nil {
			return
		}

	case 13: //官方注入NFT
		startIndex, _ := new(big.Int).SetString(w.StartIndex[2:], 16)
		if err != nil {
			return
		}
		fmt.Println("官方注入:", startIndex, w.Number, w.Royalty, w.Creator, w.Dir)
		snfts, urls, err := database.InjectSNFT(startIndex, w.Number, w.Royalty, w.Dir, w.Creator, blockNumber, timestamp)
		if err != nil {
			return
		}
		for i, url := range urls {
			//后台解析保存NFT元信息
			go SaveNFTMeta(blockNumber, snfts[i], url)
		}

	case 14: //NFT出价成交交易（卖家或交易所发起,买家给价格签名）
		nftTx := database.NFTTx{
			TxType:        2,
			NFTAddr:       w.Buyer.NFTAddress,
			ExchangerAddr: &w.Buyer.Exchanger,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 15: //NFT定价购买交易（买家发起，卖家给价格签名）
		nftTx := database.NFTTx{
			TxType:        3,
			NFTAddr:       w.Seller1.NFTAddress,
			ExchangerAddr: &w.Seller1.Exchanger,
			From:          "",     //插入数据库时实时填充原拥有者
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 16: //NFT惰性定价购买交易，买家发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return
		}
		// 获取最新NFT地址
		nftAddr, err := database.GetNFTAddr()
		if err != nil {
			return
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return
		}
		nft = &database.UserNFT{
			Address:       nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			ExchangerAddr: w.Seller2.Exchanger,
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		}
		nft.Insert()
		if err != nil {
			return
		}
		nftTx := database.NFTTx{
			TxType:        4,
			NFTAddr:       nftAddr,
			ExchangerAddr: &w.Seller2.Exchanger,
			From:          creator,
			To:            from,   //交易发起者即买家
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 17: //NFT惰性定价购买交易，交易所发起（先铸造NFT，卖家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return
		}
		// 获取最新NFT地址
		nftAddr, err := database.GetNFTAddr()
		if err != nil {
			return
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return
		}
		nft = &database.UserNFT{
			Address:       nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			ExchangerAddr: from, //交易发起者即交易所地址
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		}
		nft.Insert()
		if err != nil {
			return
		}
		nftTx := database.NFTTx{
			TxType:        5,
			NFTAddr:       nftAddr,
			ExchangerAddr: &from, //交易发起者即交易所地址
			From:          creator,
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 18: //NFT出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从授权签名恢复交易所地址
		msg := w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := recoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return
		}
		nftTx := database.NFTTx{
			TxType:        6,
			NFTAddr:       w.Buyer.NFTAddress,
			ExchangerAddr: &exchangerAddr,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 19: //NFT惰性出价成交交易，由交易所授权的地址发起（买家给价格签名）
		// 从签名恢复NFT创建者地址（也是卖家地址）
		msg := w.Seller2.Amount + w.Seller2.Royalty + w.Seller2.MetaURL + w.Seller2.ExclusiveFlag + w.Seller2.Exchanger + w.Seller2.BlockNumber
		creator, err := recoverAddress(msg, w.Seller2.Sig)
		if err != nil {
			return
		}
		// 从授权签名恢复交易所地址
		msg = w.ExchangerAuth.ExchangerOwner + w.ExchangerAuth.To + w.ExchangerAuth.BlockNumber
		exchangerAddr, err := recoverAddress(msg, w.ExchangerAuth.Sig)
		if err != nil {
			return
		}
		// 获取最新NFT地址
		nftAddr, err := database.GetNFTAddr()
		if err != nil {
			return
		}
		// 版税费率字符串转数字
		royaltyRatio, err := strconv.ParseUint(w.Seller2.Royalty[2:], 16, 32)
		if err != nil {
			return
		}
		nft = &database.UserNFT{
			Address:       nftAddr,
			RoyaltyRatio:  uint32(royaltyRatio),
			MetaUrl:       realMeatUrl(w.Seller2.MetaURL),
			ExchangerAddr: exchangerAddr, //交易发起者即交易所地址
			Creator:       creator,
			Timestamp:     timestamp,
			BlockNumber:   blockNumber,
			TxHash:        txHash,
			Owner:         creator,
		}
		nft.Insert()
		if err != nil {
			return
		}
		nftTx := database.NFTTx{
			TxType:        7,
			NFTAddr:       nftAddr,
			ExchangerAddr: &exchangerAddr,
			From:          creator,
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 20: //NFT撮合交易，交易所发起
		nftTx := database.NFTTx{
			TxType:        8,
			NFTAddr:       w.Buyer.NFTAddress,
			ExchangerAddr: &from,
			From:          "", //插入数据库时实时填充原拥有者
			To:            to,
			Price:         &value, //单位为wei
			Timestamp:     timestamp,
			TxHash:        txHash,
		}
		err = nftTx.Insert()
		if err != nil {
			return
		}

	case 21: //交易所质押
		err = database.ExchangerPledgeAdd(from, value)
		if err != nil {
			return
		}

	case 22: //撤销交易所质押
		err = database.ExchangerPledgeSub(from, value)
		if err != nil {
			return
		}
	}

	if nft != nil {
		// 后台解析
		go SaveNFTMeta(nft.BlockNumber, nft.Address, nft.MetaUrl)
	}
}

// realMeatUrl 解析真正的metaUrl
func realMeatUrl(meta string) string {
	data, err := hexutil.Decode("0x" + meta)
	if err != nil {
		return ""
	}
	real := struct {
		Meta string `json:"meta"`
	}{}
	err = json.Unmarshal(data, &real)
	if err != nil {
		return ""
	}
	return real.Meta
}

// SaveNFTMeta 解析存储NFT元信息
func SaveNFTMeta(blockNumber uint64, nftAddr, metaUrl string) {
	var err error
	defer func() {
		if err != nil {
			fmt.Println("解析存储NFT元信息失败", nftAddr, metaUrl)
		}
	}()
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}

	//合集名称+合集创建者+合集所在交易所的哈希
	var collectionId *string
	if nftMeta.CollectionsName != "" && nftMeta.CollectionsCreator != "" {
		hash := crypto.Keccak256Hash(
			[]byte(nftMeta.CollectionsName),
			[]byte(nftMeta.CollectionsCreator),
			[]byte(nftMeta.CollectionsExchanger),
		).Hex()
		collectionId = &hash
	}

	meta := &database.NFTMeta{
		NFTAddr:      nftAddr,
		Name:         nftMeta.Name,
		Desc:         nftMeta.Desc,
		Category:     nftMeta.Category,
		SourceUrl:    nftMeta.SourceUrl,
		CollectionId: collectionId,
	}
	err = meta.Insert()
	if err != nil {
		return
	}

	if collectionId != nil {
		collection := &database.Collection{
			Id:          *collectionId,
			Name:        nftMeta.CollectionsName,
			Creator:     nftMeta.CollectionsCreator,
			Category:    nftMeta.CollectionsCategory,
			Desc:        nftMeta.CollectionsDesc,
			ImgUrl:      nftMeta.CollectionsImgUrl,
			BlockNumber: blockNumber,
			Exchanger:   nftMeta.CollectionsExchanger,
		}
		err = collection.Save()
	}
}

// hashMsg Wormholes链代码复制 go-ethereum/core/evm.go 330行
func hashMsg(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

// recoverAddress Wormholes链代码复制 go-ethereum/core/evm.go 338行
func recoverAddress(msg string, sigStr string) (string, error) {
	sigData := hexutil.MustDecode(sigStr)
	if len(sigData) != 65 {
		return common.Address{}.Hex(), fmt.Errorf("signature must be 65 bytes long")
	}
	if sigData[64] != 27 && sigData[64] != 28 {
		return common.Address{}.Hex(), fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sigData[64] -= 27
	hash, _ := hashMsg([]byte(msg))
	rpk, err := crypto.SigToPub(hash, sigData)
	if err != nil {
		return common.Address{}.Hex(), err
	}
	return strings.ToLower(crypto.PubkeyToAddress(*rpk).Hex()), nil
}

func toDecimal(src string, decimal int) string {
	var ret string
	if len(src) < decimal {
		ret += "0."
		for i := 0; i < decimal-len(src); i++ {
			ret += "0"
		}
		ret += src
		return ret
	} else {
		preLen := len(src) - decimal
		if preLen == 0 {
			return string(src[0]) + "." + src[1:]
		}
		return src[0:preLen] + "." + src[preLen:]
	}
}
