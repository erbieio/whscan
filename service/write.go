package service

import (
	"fmt"
	"math/big"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/types"
	"server/common/utils"
	"server/ethclient"
	"server/model"
)

// cache 缓存一些数据库查询，加速查询
var cache = struct {
	TotalBlock       uint64 //总区块数量
	TotalTransaction uint64 //总交易数量
	TotalUncle       uint64 //总叔块数量
	TotalUserNFT     uint64 //用户NFT总数
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
	if err = DB.Model(&model.UserNFT{}).Select("COUNT(*)").Scan(&number).Error; err != nil {
		return err
	}
	cache.TotalUserNFT = number
	return err
}

func TotalBlock() uint64 {
	return cache.TotalBlock
}

func TotalTransaction() uint64 {
	return cache.TotalTransaction
}

// getNFTAddr 获取NFT地址
func getNFTAddr(next *big.Int) string {
	return string(utils.BigToAddress(next.Add(next, big.NewInt(int64(cache.TotalUserNFT+1)))))
}

func BlockInsert(block *ethclient.DecodeRet) error {
	err := DB.Transaction(func(t *gorm.DB) error {
		for _, log := range block.CacheLogs {
			if len(log.Topics) > 0 {
				log.EventID = (*types.Hash)(&log.Topics[0])
			}
			// 写入交易日志
			if err := t.Create(log).Error; err != nil {
				return err
			}
			// 解析写入ERC合约转移事件
			erc, err := erc(log.Address)
			if err != nil {
				return err
			}
			switch erc {
			case types.ERC20:
				if transferLog, err := utils.Unpack20TransferLog(parseLog(log)); err == nil {
					err = t.Create(&model.ERC20Transfer{
						TxHash:  log.TxHash,
						Address: log.Address,
						From:    types.Address(strings.ToLower(transferLog.From.Hex())),
						To:      types.Address(strings.ToLower(transferLog.To.Hex())),
						Value:   types.BigInt(transferLog.Value.String()),
					}).Error
					if err != nil {
						return err
					}
				}
			case types.ERC721:
				if transferLog, err := utils.Unpack721TransferLog(parseLog(log)); err == nil {
					err = t.Create(&model.ERC721Transfer{
						TxHash:  log.TxHash,
						Address: log.Address,
						From:    types.Address(strings.ToLower(transferLog.From.Hex())),
						To:      types.Address(strings.ToLower(transferLog.To.Hex())),
						TokenId: types.BigInt(transferLog.TokenId.String()),
					}).Error
					if err != nil {
						return err
					}
				}
			case types.ERC1155:
				if transferLog, err := utils.Unpack1155TransferSingleLog(parseLog(log)); err == nil {
					err = t.Create(&model.ERC1155Transfer{
						TxHash:  log.TxHash,
						Address: log.Address,
						From:    types.Address(strings.ToLower(transferLog.From.Hex())),
						To:      types.Address(strings.ToLower(transferLog.To.Hex())),
						TokenId: types.BigInt(transferLog.Id.String()),
						Value:   types.BigInt(transferLog.Value.String()),
					}).Error
					if err != nil {
						return err
					}
				} else {
					if transferBatchLog, err := utils.Unpack1155TransferBatchLog(parseLog(log)); err == nil {
						from := types.Address(strings.ToLower(transferBatchLog.From.Hex()))
						to := types.Address(strings.ToLower(transferBatchLog.To.Hex()))
						for i := range transferBatchLog.Ids {
							err = t.Create(&model.ERC1155Transfer{
								TxHash:  log.TxHash,
								Address: log.Address,
								From:    from,
								To:      to,
								TokenId: types.BigInt(transferBatchLog.Ids[i].String()),
								Value:   types.BigInt(transferBatchLog.Values[i].String()),
							}).Error
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

		for _, tx := range block.CacheTxs {
			// 写入区块交易
			if len(tx.Input) > 10 {
				MethodId := tx.Input[:10]
				tx.MethodId = (*types.Bytes8)(&MethodId)
			}
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

		// 写入区块
		if err := t.Create(block.Block).Error; err != nil {
			return err
		}
		return WHInsert(t, block)
	})

	// 如果写入数据库成功，则更新查询缓存
	if err == nil {
		cache.TotalBlock++
		cache.TotalTransaction += uint64(block.TotalTransaction)
		cache.TotalUncle += uint64(block.UnclesCount)
		cache.TotalUserNFT += uint64(len(block.CreateNFTs))
	}
	return err
}

func erc(addr types.Address) (erc types.ERC, err error) {
	contract := model.Contract{Address: addr}
	err = DB.Find(&contract).Error
	erc = contract.ERC
	return
}

func WHInsert(tx *gorm.DB, wh *ethclient.DecodeRet) (err error) {
	// SNFT创建
	for _, snft := range wh.CreateSNFTs {
		err = tx.Create(snft).Error
		if err != nil {
			return
		}
		go SaveNFTMeta(uint64(wh.Number), snft.Address, snft.MetaUrl)
	}
	// SNFT奖励
	for _, snft := range wh.RewardSNFTs {
		err = tx.Select("owner", "awardee", "reward_number", "reward_at").Save(snft).Error
		if err != nil {
			return
		}
	}
	// UserNFT创建
	for i, nft := range wh.CreateNFTs {
		*nft.Address = getNFTAddr(big.NewInt(int64(i)))
		err = tx.Create(nft).Error
		if err != nil {
			return
		}
		go SaveNFTMeta(uint64(wh.Number), *nft.Address, nft.MetaUrl)
	}
	// NFT交易，包含用户和官方类型的NFT
	for _, nftTx := range wh.NFTTxs {
		err = SaveNFTTx(tx, nftTx)
		if err != nil {
			return
		}
	}
	// 交易所创建或者关闭
	for _, exchanger := range wh.Exchanger {
		err = SaveExchanger(tx, exchanger)
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
	// 官方注入SNFT元信息
	for _, snft := range wh.InjectSNFTs {
		err = InjectSNFT(tx, snft.StartIndex, snft.Count, snft.Royalty, snft.Dir, snft.Creator, snft.Number, snft.Timestamp)
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

// SaveNFTMeta 解析存储NFT元信息
func SaveNFTMeta(blockNumber uint64, nftAddr, metaUrl string) {
	var err error
	defer func() {
		if err != nil {
			fmt.Println("解析存储NFT元信息失败", nftAddr, metaUrl, err)
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
			BlockNumber: blockNumber,
			Exchanger:   nftMeta.CollectionsExchanger,
		}).Error
	}
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
		err = tx.Model(&model.UserNFT{}).Where("address=?", nft.Address).UpdateColumns(map[string]interface{}{
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
		err = tx.Model(&model.OfficialNFT{}).Where("address=?", nft.Address).UpdateColumns(map[string]interface{}{
			"last_price": nt.Price,
			"owner":      nt.To,
		}).Error
	}
	return tx.Create(&nt).Error
}

// SaveExchanger 更新交易所信息
func SaveExchanger(tx *gorm.DB, e *model.Exchanger) error {
	if e.IsOpen {
		err := tx.Where("address=?", e.Address).First(&model.Exchanger{}).Error
		if err == gorm.ErrRecordNotFound {
			return tx.Create(e).Error
		}
		return tx.Model(&model.Exchanger{}).Where("address=?", e.Address).UpdateColumns(map[string]interface{}{
			"is_open":   true,
			"name":      e.Name,
			"url":       e.URL,
			"fee_ratio": e.FeeRatio,
		}).Error
	} else {
		return tx.Model(&model.Exchanger{}).Where("address=?", e.Address).Update("is_open", false).Error
	}
}

// RecycleSNFT SNFT回收,清空所有者
func RecycleSNFT(tx *gorm.DB, addr string) error {
	if len(addr) == 42 {
		return tx.Model(&model.OfficialNFT{}).Where("address=?", addr).UpdateColumns(map[string]interface{}{
			"owner":         nil,
			"Awardee":       nil,
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
	return tx.Create(&pledge).Error
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
	return tx.Create(&pledge).Error
}
