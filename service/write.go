package service

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// cache 缓存一些数据库查询，加速查询
var cache = struct {
	TotalBlock       uint64       //总区块数量
	TotalTransaction uint64       //总交易数量
	TotalUncle       uint64       //总叔块数量
	TotalAccount     uint64       //总账户数量
	TotalBalance     types.BigInt //链的币总额
	TotalUserNFT     uint64       //用户NFT总数
	TotalOfficialNFT uint64       //官方NFT总数
}{}

// InitCache 从数据库初始化查询缓存
func initCache() (err error) {
	var number uint64
	if err = DB.Model(&model.Block{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalBlock = number
	if err = DB.Model(&model.Transaction{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalTransaction = number
	if err = DB.Model(&model.Uncle{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalUncle = number
	if err = DB.Model(&model.Account{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalAccount = number
	// todo 计算和缓存币总额
	cache.TotalBalance = "1000000000000000000000000000000000000000000000000000000000000000"
	if err = DB.Model(&model.UserNFT{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalUserNFT = number
	if err = DB.Model(&model.OfficialNFT{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalOfficialNFT = number
	return err
}

func TotalBlock() uint64 {
	return cache.TotalBlock
}

func TotalTransaction() uint64 {
	return cache.TotalTransaction
}

func TotalOfficialNFT() uint64 {
	return cache.TotalOfficialNFT
}

var lastAccount = time.Now()

func TotalAccount() uint64 {
	if time.Now().Sub(lastAccount).Seconds() > 60 {
		var number uint64
		if err := DB.Model(&model.Account{}).Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.TotalAccount = number
		}
	}
	return cache.TotalAccount
}

func TotalBalance() types.BigInt {
	return cache.TotalBalance
}

// getNFTAddr 获取NFT地址
func getNFTAddr(next *big.Int) string {
	return string(utils.BigToAddress(next.Add(next, big.NewInt(int64(cache.TotalUserNFT+1)))))
}

// DecodeRet 区块解析结果
type DecodeRet struct {
	*model.Block
	CacheTxs         []*model.Transaction `json:"transactions"`
	CacheInternalTxs []*model.InternalTx
	CacheUncles      []*model.Uncle
	CacheAccounts    map[types.Address]*model.Account
	CacheContracts   map[types.Address]*model.Contract
	CacheLogs        []*model.Log //在CacheContracts之后插入

	// wormholes，需要按优先级插入数据库（后面的数据可能会查询先前数据）
	Exchangers       []*model.Exchanger       //交易所,优先级：1
	InjectSNFTs      []*Inject                //官方注入SNFT,优先级：1
	RecycleSNFTs     []string                 //回收的SNFT,优先级：1
	CreateSNFTs      []*model.OfficialNFT     //SNFT的创建信息,优先级：2
	CreateNFTs       []*model.UserNFT         //新创建的NFT,优先级：2
	RewardSNFTs      []*model.OfficialNFT     //SNFT的奖励信息,优先级：3
	NFTTxs           []*model.NFTTx           //NFT交易记录,优先级：4
	ExchangerPledges []*model.ExchangerPledge //交易所质押,优先级：无
	ConsensusPledges []*model.ConsensusPledge //共识质押,优先级：无
}

type Inject struct {
	StartIndex        *big.Int
	Count             uint64
	Royalty           uint32
	Dir, Creator      string
	Number, Timestamp uint64
}

func BlockInsert(block *DecodeRet) error {
	err := DB.Transaction(func(t *gorm.DB) error {
		for _, tx := range block.CacheTxs {
			// 写入区块交易
			if err := t.Create(tx).Error; err != nil {
				return err
			}
		}

		for _, internalTx := range block.CacheInternalTxs {
			// 写入内部交易
			if err := t.Create(internalTx).Error; err != nil {
				return err
			}
		}

		for _, uncle := range block.CacheUncles {
			// 写入叔块
			if err := t.Create(uncle).Error; err != nil {
				return err
			}
		}

		for _, account := range block.CacheAccounts {
			// 写入账户信息
			if err := t.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"balance", "nonce", "code_hash"}),
			}).Create(account).Error; err != nil {
				return err
			}
		}

		for _, contract := range block.CacheContracts {
			// 写入合约信息
			if err := t.Clauses(clause.OnConflict{UpdateAll: true}).Create(contract).Error; err != nil {
				return err
			}
		}

		// 写入交易日志
		err := SaveTxLog(t, block.CacheLogs)
		if err != nil {
			return err
		}

		// 写入区块
		if err := t.Create(block.Block).Error; err != nil {
			return err
		}
		// wormholes独有数据写入
		return WHInsert(t, block)
	})

	// 如果写入数据库成功，则更新查询缓存
	if err == nil {
		cache.TotalBlock++
		cache.TotalTransaction += uint64(block.TotalTransaction)
		cache.TotalUncle += uint64(block.UnclesCount)
		cache.TotalUserNFT += uint64(len(block.CreateNFTs))
		cache.TotalOfficialNFT += uint64(len(block.CreateSNFTs)) //todo 可能存在误差，注入和销毁的情况
	}
	return err
}

// SaveTxLog 写入区块交易日志，并分析存储ERC代币交易
func SaveTxLog(tx *gorm.DB, cacheLog []*model.Log) error {
	for _, cacheLog := range cacheLog {
		// 写入交易日志
		if err := tx.Create(cacheLog).Error; err != nil {
			return err
		}
		// 解析写入ERC合约转移事件
		contract := model.Contract{Address: cacheLog.Address}
		err := DB.Find(&contract).Error
		if err != nil {
			return err
		}
		switch contract.ERC {
		case types.ERC20:
			if transferLog, err := utils.Unpack20TransferLog(cacheLog); err == nil {
				err = tx.Create(transferLog).Error
				if err != nil {
					return err
				}
			}
		case types.ERC721:
			if transferLog, err := utils.Unpack721TransferLog(cacheLog); err == nil {
				err = tx.Create(transferLog).Error
				if err != nil {
					return err
				}
			}
		case types.ERC1155:
			if transferLogs, err := utils.Unpack1155TransferLog(cacheLog); err == nil {
				err = tx.Create(transferLogs).Error
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func WHInsert(tx *gorm.DB, wh *DecodeRet) (err error) {
	// 交易所创建或者关闭
	for _, exchanger := range wh.Exchangers {
		err = SaveExchanger(tx, exchanger)
		if err != nil {
			return
		}
	}
	// 官方注入SNFT元信息
	for _, snft := range wh.InjectSNFTs {
		err = InjectSNFT(tx, snft.StartIndex, snft.Count, snft.Royalty, snft.Dir, snft.Creator, snft.Number, snft.Timestamp)
		if err != nil {
			return
		}
	}
	// 回收SNFT
	for _, snft := range wh.RecycleSNFTs {
		err = RecycleSNFT(tx, snft)
		if err != nil {
			return
		}
	}
	// SNFT创建
	for _, snft := range wh.CreateSNFTs {
		err = tx.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(snft).Error
		if err != nil {
			return
		}
		go SaveNFTMeta(wh.Number, snft.Address, snft.MetaUrl)
	}
	// SNFT奖励
	for _, snft := range wh.RewardSNFTs {
		err = tx.Select("owner", "awardee", "reward_number", "reward_at").Updates(snft).Error
		if err != nil {
			return
		}
	}
	// UserNFT创建
	err = SaveUserNFT(tx, wh.Number, wh.CreateNFTs)
	if err != nil {
		return
	}
	// NFT交易，包含用户和官方类型的NFT
	for _, nftTx := range wh.NFTTxs {
		err = SaveNFTTx(tx, nftTx)
		if err != nil {
			return
		}
	}
	// 交易所质押
	for _, pledge := range wh.ExchangerPledges {
		err = ExchangerPledgeAdd(tx, pledge.Address, pledge.Amount)
		if err != nil {
			return
		}
	}
	// 共识质押
	for _, pledge := range wh.ConsensusPledges {
		err = ConsensusPledgeAdd(tx, pledge.Address, pledge.Amount)
		if err != nil {
			return
		}
	}
	return
}

// SaveExchanger 更新交易所信息
func SaveExchanger(tx *gorm.DB, e *model.Exchanger) error {
	if e.IsOpen {
		err := tx.Where("address=?", e.Address).First(&model.Exchanger{}).Error
		if err == gorm.ErrRecordNotFound {
			return tx.Create(e).Error
		}
		return tx.Model(&model.Exchanger{}).Where("address=?", e.Address).Updates(map[string]interface{}{
			"is_open":   true,
			"name":      e.Name,
			"url":       e.URL,
			"fee_ratio": e.FeeRatio,
		}).Error
	} else {
		return tx.Model(&model.Exchanger{}).Where("address=?", e.Address).Update("is_open", false).Error
	}
}

// SaveUserNFT 保存用户创建的NFT
func SaveUserNFT(tx *gorm.DB, number types.Uint64, nfts []*model.UserNFT) error {
	for i, nft := range nfts {
		*nft.Address = getNFTAddr(big.NewInt(int64(i)))
		err := tx.Create(nft).Error
		if err != nil {
			return err
		}
		// 更新指定交易所的总NFT数
		if nft.ExchangerAddr != "" {
			var exchanger model.Exchanger
			err := tx.Find(&exchanger, "address=?", nft.ExchangerAddr).Error
			if err != nil {
				return err
			}
			// todo 可能交易所不存在
			if exchanger.Address == nft.ExchangerAddr {
				exchanger.NFTCount++
				err = tx.Select("nft_count").Updates(&exchanger).Error
				if err != nil {
					return err
				}
			}
		}
		go SaveNFTMeta(number, *nft.Address, nft.MetaUrl)
	}
	return nil
}

// SaveNFTTx 保存NFT交易记录
func SaveNFTTx(tx *gorm.DB, nt *model.NFTTx) error {
	// 更新NFT所有者和最新价格
	if (*nt.NFTAddr)[2] != '8' {
		var nft model.UserNFT
		err := tx.Where("address=?", nt.NFTAddr).First(&nft).Error
		if err != nil {
			return err
		}
		// 填充卖家字段（如果没有）
		if nt.From == "" {
			nt.From = nft.Owner
		}
		err = tx.Model(&model.UserNFT{}).Where("address=?", nft.Address).Updates(map[string]interface{}{
			"last_price": nt.Price,
			"owner":      nt.To,
		}).Error
	} else {
		// todo 处理合成的SNFT碎片地址
		var nft model.OfficialNFT
		err := tx.Where("address=?", nt.NFTAddr).First(&nft).Error
		if err != nil {
			return err
		}
		// 填充卖家字段（如果没有）
		if nt.From == "" {
			nt.From = *nft.Owner
		}
		err = tx.Model(&model.OfficialNFT{}).Where("address=?", nft.Address).Updates(map[string]interface{}{
			"last_price": nt.Price,
			"owner":      nt.To,
		}).Error
	}
	// 计算填充NFT交易手续费和保存交易所的总交易数和总交易额
	if nt.ExchangerAddr != "" && nt.Price != nil && *nt.Price != "0" {
		var exchanger model.Exchanger
		err := tx.Find(&exchanger, "address=?", nt.ExchangerAddr).Error
		if err != nil {
			return err
		}
		// todo 可能交易所不存在
		if exchanger.Address == nt.ExchangerAddr && exchanger.FeeRatio > 0 {
			price, _ := new(big.Int).SetString(*nt.Price, 10)
			fee := big.NewInt(int64(exchanger.FeeRatio))
			fee = fee.Mul(fee, price)
			feeStr := fee.Div(fee, big.NewInt(10000)).Text(10)
			nt.Fee = &feeStr
			exchanger.TxCount++
			balanceCount := new(big.Int)
			balanceCount.SetString(exchanger.BalanceCount, 10)
			balanceCount = balanceCount.Add(balanceCount, price)
			exchanger.BalanceCount = balanceCount.Text(10)
			err = tx.Select("tx_count", "balance_count").Updates(&exchanger).Error
			if err != nil {
				return err
			}
		}
	}
	return tx.Create(&nt).Error
}

// RecycleSNFT SNFT回收,清空所有者
func RecycleSNFT(tx *gorm.DB, addr string) error {
	if len(addr) == 42 {
		return tx.Model(&model.OfficialNFT{}).Where("address=?", addr).Updates(map[string]interface{}{
			"owner":         nil,
			"awardee":       nil,
			"reward_at":     nil,
			"reward_number": nil,
		}).Error
	} else {
		for i := 0; i < 16; i++ {
			err := RecycleSNFT(tx, fmt.Sprintf("%s%x", addr, i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// InjectSNFT 官方批量注入SNFT
func InjectSNFT(tx *gorm.DB, startIndex *big.Int, count uint64, royalty uint32, dir, creator string, number, timestamp uint64) (err error) {
	SNFTAddr := big.NewInt(0)
	Big1 := big.NewInt(1)
	SNFTAddr.SetString("8000000000000000000000000000000000000000", 16)
	SNFTAddr = SNFTAddr.Add(SNFTAddr, startIndex)
	for i := uint64(0); i < count; i++ {
		addr := string(utils.BigToAddress(SNFTAddr))
		// 取地址倒数3-4位作为文件名
		metaUrl := dir + addr[39:40]
		err = tx.Create(&model.OfficialNFT{
			Address:      strings.ToLower(addr),
			CreateAt:     timestamp,
			CreateNumber: number,
			Creator:      creator,
			RoyaltyRatio: royalty,
			MetaUrl:      metaUrl,
		}).Error
		if err != nil {
			return err
		}
		SNFTAddr = SNFTAddr.Add(SNFTAddr, Big1)
	}
	return
}

// ExchangerPledgeAdd 增加质押金额(amount为负数则减少)
func ExchangerPledgeAdd(tx *gorm.DB, addr, amount string) error {
	pledge := model.ExchangerPledge{}
	err := tx.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		pledge.Amount = "0x0"
	}
	pledge.Amount = BigIntAdd(pledge.Amount, amount)
	return tx.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"amount", "count"}),
	}).Create(&pledge).Error
}

// ConsensusPledgeAdd 增加质押金额(amount为负数则减少)
func ConsensusPledgeAdd(tx *gorm.DB, addr, amount string) error {
	pledge := model.ConsensusPledge{}
	err := tx.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		pledge.Amount = "0x0"
	}
	pledge.Amount = BigIntAdd(pledge.Amount, amount)
	return tx.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"amount", "count"}),
	}).Create(&pledge).Error
}

// SaveNFTMeta 解析存储NFT元信息
func SaveNFTMeta(blockNumber types.Uint64, nftAddr, metaUrl string) {
	var err error
	defer func() {
		if err != nil {
			log.Println("解析存储NFT元信息失败", nftAddr, metaUrl, err)
		}
	}()
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}

	//合集名称+合集创建者+合集所在交易所的哈希
	var collectionId *string
	if nftMeta.CollectionsName != "" && nftMeta.CollectionsCreator != "" {
		hash := string(utils.Keccak256Hash(
			[]byte(nftMeta.CollectionsName),
			[]byte(nftMeta.CollectionsCreator),
			[]byte(nftMeta.CollectionsExchanger),
		))
		collectionId = &hash
	}

	err = DB.Create(&model.NFTMeta{
		NFTAddr:      nftAddr,
		Name:         nftMeta.Name,
		Desc:         nftMeta.Desc,
		Category:     nftMeta.Category,
		SourceUrl:    nftMeta.SourceUrl,
		CollectionId: collectionId,
	}).Error
	if err == nil && collectionId != nil {
		err = DB.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&model.Collection{
			Id:          *collectionId,
			Name:        nftMeta.CollectionsName,
			Creator:     nftMeta.CollectionsCreator,
			Category:    nftMeta.CollectionsCategory,
			Desc:        nftMeta.CollectionsDesc,
			ImgUrl:      nftMeta.CollectionsImgUrl,
			BlockNumber: uint64(blockNumber),
			Exchanger:   nftMeta.CollectionsExchanger,
		}).Error
	}
}
