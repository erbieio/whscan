package service

import (
	"encoding/json"
	"log"
	"math/big"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

func Insert(parsed *model.Parsed) (head types.Long, err error) {
	err = DB.Transaction(func(db *gorm.DB) (err error) {
		err = db.Model(&model.Block{}).Where("`hash`=?", parsed.ParentHash).Select("`number`+1").Scan(&head).Error
		if err != nil || parsed.Number != head {
			return
		}
		// write block transaction
		if len(parsed.CacheTxs) > 0 {
			if err = db.Create(parsed.CacheTxs).Error; err != nil {
				return
			}
		}
		// write transaction cacheLog
		if len(parsed.CacheLogs) > 0 {
			if err = db.Create(parsed.CacheLogs).Error; err != nil {
				return
			}
		}
		// write uncle block
		if len(parsed.CacheUncles) > 0 {
			if err = db.Create(parsed.CacheUncles).Error; err != nil {
				return
			}
		}
		// write internal transaction
		if len(parsed.CacheInternalTxs) > 0 {
			if err = db.Create(parsed.CacheInternalTxs).Error; err != nil {
				return
			}
		}
		// write transaction transferLog
		for _, cacheTransferLog := range parsed.CacheTransferLogs {
			if err = db.Create(cacheTransferLog).Error; err != nil {
				return
			}
		}
		// write account information
		if len(parsed.CacheAccounts) > 0 {
			if err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"balance", "nonce", "number", "snft_value"}),
			}).Create(parsed.CacheAccounts).Error; err != nil {
				return
			}
		}
		// write block
		if err = db.Create(parsed.Block).Error; err != nil {
			return
		}
		// write collection
		if len(parsed.Collections) > 0 {
			if err = db.Save(parsed.Collections).Error; err != nil {
				return
			}
		}

		// wormholes unique data write
		// NFT creation
		if err = saveNFT(db, parsed); err != nil {
			return
		}
		// NFT transactions, including user and official types of NFTs
		if err = saveNFTTx(db, parsed); err != nil {
			return
		}
		// validator change
		if err = saveValidator(db, parsed); err != nil {
			return
		}
		// exchanger change
		if err = saveExchanger(db, parsed); err != nil {
			return
		}
		// Officially inject SNFT meta information
		if err = injectSNFT(db, parsed); err != nil {
			return
		}
		// reward save
		if err = saveReward(db, parsed); err != nil {
			return
		}
		if err = saveMerge(db, parsed); err != nil {
			return
		}

		// update the query stats
		return updateStats(db, parsed)
	})
	freshStats(DB, parsed)
	if parsed.Number%17280 == 0 {
		// once a day
		reTryMeta(DB)
	}
	return
}

func VerifyHead(parsed *model.Parsed) (pass bool, err error) {
	err = DB.Model(&model.Block{}).Where("`hash`=? AND `number`=?", parsed.Hash, parsed.Number).Select("COUNT(*)=1").Scan(&pass).Error
	if pass {
		err = DB.Select("address").Find(&parsed.CacheAccounts, "`number`>?", parsed.Number).Error
	}
	var errBlock model.Block
	if DB.Find(&errBlock, "`number`=?", parsed.Number).Error == nil {
		errB, _ := json.Marshal(errBlock.Header)
		okB, _ := json.Marshal(parsed.Header)
		log.Printf("err block: %+v expect: %+v", string(errB), string(okB))
	}
	return
}

func SetHead(parsed *model.Parsed) error {
	return DB.Transaction(func(db *gorm.DB) (err error) {
		if head := parsed.Number; head >= 0 {
			if err = db.Delete(&model.FNFT{}, "LEFT(`id`, 39) IN (?)",
				db.Model(&model.Epoch{}).Select("id").Where("number>?", head),
			).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Epoch{}, "number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Collection{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Exchanger{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Model(&model.Exchanger{}).Where("close_at>?", head).Update("close_at", nil).Error; err != nil {
				return
			}
			hashes := db.Model(&model.Transaction{}).Select("hash").Where("block_number>?", head)
			if err = db.Delete(&model.Account{}, "created_tx IN (?)", hashes).Error; err != nil {
				return
			}
			if err = db.Delete(&model.ERC20Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = db.Delete(&model.ERC721Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = db.Delete(&model.ERC1155Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = db.Delete(&model.InternalTx{}, "`tx_hash` IN (?)", hashes).Error; err != nil {
				return
			}
			if err = db.Delete(&model.EventLog{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.NFTTx{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.NFT{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.SNFT{}, "reward_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Reward{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Transaction{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Block{}, "number>?", head).Error; err != nil {
				return
			}
			return fixStats(db, parsed)
		} else {
			if err = model.ClearTable(db); err != nil {
				return
			}
			return initStats(db)
		}
	})
}

// injectSNFT official batch injection of SNFT
func injectSNFT(db *gorm.DB, wh *model.Parsed) (err error) {
	if epoch := wh.Epoch; epoch != nil {
		// full txHash, txType and timestamp
		var tx *struct {
			Input string     `json:"input"`
			Hash  types.Hash `json:"hash" gorm:"type:CHAR(66);primaryKey"`
		}
		err = db.Model(&model.Transaction{}).Where("block_number>? AND block_number<=? AND status=1", epoch.Number-1024, epoch.Number).
			Where("LEFT(input,76)='0x776f726d686f6c65733a7b2276657273696f6e223a22302e3031222c2274797065223a3233' OR LEFT(input,76)='0x776f726d686f6c65733a7b2276657273696f6e223a22302e3031222c2274797065223a3234'").
			Order("block_number DESC").Limit(1).Scan(&tx).Error
		if err != nil {
			return
		}
		if tx != nil {
			epoch.TxHash = (*string)(&tx.Hash)
			epoch.TxType = new(int64)
			*epoch.TxType = -28 + int64(tx.Input[75])
		}
		err = db.Model(&Block{}).Where("number=?", epoch.Number).Select("timestamp").Scan(&epoch.Timestamp).Error
		if err != nil {
			return
		}

		// full weightAmount and update creator
		creator := model.Creator{}
		err = db.Find(&creator, "address=?", epoch.Creator).Error
		if err != nil {
			return
		}
		if creator.Address != "" {
			epoch.WeightAmount = epoch.Number - creator.LastNumber
			creator.LastNumber = epoch.Number
			creator.LastTime = epoch.Timestamp
			creator.LastEpoch = epoch.ID
			creator.Count++
		} else {
			creator.Address = epoch.Creator
			creator.Number = epoch.Number
			creator.Timestamp = epoch.Timestamp
			creator.LastEpoch = epoch.ID
			creator.LastNumber = epoch.Number
			creator.LastTime = epoch.Timestamp
			creator.Count = 1
			creator.Reward = "0"
			creator.Profit = "0"
		}
		if epoch.Creator == epoch.Voter {
			creator.Reward = BigIntAdd(creator.Reward, epoch.Reward)
		}
		if err = db.Save(&creator).Error; err != nil {
			return
		}
		if err = db.Create(epoch).Error; err != nil {
			return
		}
		if err = db.Create(wh.FNFTs).Error; err != nil {
			return
		}
	}
	return
}

func saveMerge(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, snft := range wh.Mergers {
		addr := [16]string{}
		for i := int64(0); i < 16; i++ {
			addr[i] = snft.Address + strconv.FormatInt(i, 16)
		}
		if err = db.Model(&model.SNFT{}).Where("address IN (?)", addr).Update("remove", true).Error; err != nil {
			return
		}
		if err = db.Create(snft).Error; err != nil {
			return
		}
	}
	return
}

func saveReward(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, reward := range wh.Rewards {
		err = db.Exec("INSERT INTO rewards (`address`, `proxy`, `identity`, `block_number`, `snft`, `amount`) VALUES "+
			"(@Address, (SELECT `proxy` FROM validators WHERE address=@Address), @Identity, @BlockNumber, @SNFT, @Amount)", reward).Error
		if err != nil {
			return
		}
		if reward.SNFT != nil {
			err = db.Create(&model.SNFT{
				Address:      *reward.SNFT,
				TxAmount:     "0",
				RewardAt:     int64(wh.Timestamp),
				RewardNumber: int64(wh.Number),
				Owner:        reward.Address,
				Pieces:       1,
			}).Error
			if err != nil {
				return
			}
			var exchanger model.Exchanger
			if err = db.Find(&exchanger, "`amount`!='0' AND `address`=?", reward.Address).Error; err != nil {
				return
			}
			value := snftValue(*reward.SNFT, 1)
			if exchanger.Address != "" {
				exchanger.Reward = BigIntAdd(exchanger.Reward, value)
				exchanger.RewardCount++
				if err = db.Select("reward", "reward_count").Updates(&exchanger).Error; err != nil {
					return
				}
			}
		} else {
			var validator model.Validator
			if err = db.Find(&validator, "address=?", reward.Address).Error; err != nil {
				return
			}
			if validator.Address != "" {
				validator.Reward = BigIntAdd(validator.Reward, *reward.Amount)
				validator.RewardCount++
				validator.Timestamp = int64(wh.Timestamp)
				validator.LastNumber = int64(wh.Number)
				if err = db.Select("reward", "reward_count", "timestamp", "last_number").Updates(&validator).Error; err != nil {
					return
				}
			}
		}
	}
	return
}

// saveNFT saves the NFT created by the user
func saveNFT(db *gorm.DB, wh *model.Parsed) (err error) {
	for i, nft := range wh.NFTs {
		*nft.Address = string(utils.BigToAddress(big.NewInt(int64(i) + stats.TotalNFT + 1)))
		if err = db.Create(nft).Error; err != nil {
			return
		}
		// Update the total number of NFTs on the specified exchange
		if len(nft.ExchangerAddr) == 42 {
			if err = db.Exec("UPDATE `exchangers` SET `nft_count`=`nft_count`+1 WHERE address=?", nft.ExchangerAddr).Error; err != nil {
				return
			}
		}
	}
	return
}

// saveNFTTx saves the NFT transaction record, while updating the NFT owner and the latest price
func saveNFTTx(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, tx := range wh.NFTTxs {
		if tx.TxType == 6 {
			// handle recycle snft
			var snft model.SNFT
			if err = db.Where("address=?", tx.NFTAddr).Take(&snft).Error; err != nil {
				return
			}
			fee := strconv.Itoa(int(snft.Pieces))
			tx.Fee = &fee
			tx.Price = snftValue(snft.Address, snft.Pieces)
			if err = db.Model(&model.SNFT{}).Where("address=?", snft.Address).Update("remove", true).Error; err != nil {
				return
			}
		} else {
			if (*tx.NFTAddr)[2] != '8' {
				// handle NFT
				var nft model.NFT
				if err = db.Where("address=?", tx.NFTAddr).Take(&nft).Error; err != nil {
					return
				}
				// populate seller field (if none)
				if tx.From == "" {
					tx.From = nft.Owner
				}
				tx.Royalty = "0"
				if tx.Price != "0" {
					tx.Royalty = *TxFee(tx.Price, nft.RoyaltyRatio)
					nft.LastPrice = &tx.Price
					nft.TxAmount = BigIntAdd(nft.TxAmount, tx.Price)
				}
				nft.Owner = tx.To
				if err = db.Select("last_price", "tx_amount", "owner").Updates(&nft).Error; err != nil {
					return
				}
			} else if tx.TxType == 28 {
				addr := [16]string{}
				for i := int64(0); i < 16; i++ {
					addr[i] = *tx.NFTAddr + strconv.FormatInt(i, 16)
				}
				result := db.Model(&model.SNFT{}).Where("address IN (?) AND owner!=?", addr, tx.To).Update("owner", tx.To)
				if err = result.Error; err != nil {
					return
				}
				tx.Price = snftValue(*tx.NFTAddr, result.RowsAffected)
				tx.Royalty = *TxFee(tx.Price, 1000)
			} else {
				var snft model.SNFT
				if err = db.Where("address=?", tx.NFTAddr).Take(&snft).Error; err != nil {
					return
				}
				// populate seller field (if none)
				if tx.From == "" {
					tx.From = snft.Owner
				}
				tx.Royalty = "0"
				if tx.Price != "0" {
					tx.Royalty = *TxFee(tx.Price, 1000)
					snft.LastPrice = &tx.Price
					snft.TxAmount = BigIntAdd(snft.TxAmount, tx.Price)
				}
				snft.Owner = tx.To
				if err = db.Select("last_price", "tx_amount", "owner").Updates(&snft).Error; err != nil {
					return
				}
			}
			if (*tx.NFTAddr)[2] == '8' && tx.Royalty != "0" {
				creator := model.Creator{}
				err = db.Find(&creator, "address=(?)", db.Model(&model.Epoch{}).Where("`id`=?", (*tx.NFTAddr)[:39]).Select("creator")).Error
				if err != nil {
					return
				}
				creator.Profit = BigIntAdd(creator.Profit, tx.Royalty)
				if err = db.Select("profit").Updates(creator).Error; err != nil {
					return
				}
				epoch := model.Epoch{}
				err = db.Find(&epoch, "`id`=(?)", (*tx.NFTAddr)[:39]).Error
				if err != nil {
					return
				}
				epoch.Profit = BigIntAdd(epoch.Profit, tx.Royalty)
				if err = db.Select("profit").Updates(epoch).Error; err != nil {
					return
				}
			}
			// Calculate the total number of transactions and the total transaction amount to fill the NFT transaction fee and save the exchange
			if tx.ExchangerAddr != nil && tx.Price != "0" {
				var exchanger model.Exchanger
				db.Find(&exchanger, "address=?", tx.ExchangerAddr)
				if exchanger.Address != "" {
					tx.Fee = TxFee(tx.Price, exchanger.FeeRatio)
					exchanger.TxAmount = BigIntAdd(exchanger.TxAmount, tx.Price)
					if err = db.Select("tx_amount").Updates(&exchanger).Error; err != nil {
						return
					}
				}
			}
		}
		if err = db.Create(&tx).Error; err != nil {
			return
		}
	}
	return
}

func saveValidator(db *gorm.DB, wh *model.Parsed) (err error) {
	for i, change := range wh.ChangeValidators {
		if wh.Number > 0 && i < 11 {
			if err = db.Select("weight").Updates(change).Error; err != nil {
				return
			}
		} else {
			validator := model.Validator{Amount: "0", Reward: "0", Weight: 70}
			if err = db.Find(&validator, "address=?", change.Address).Error; err != nil {
				return
			}
			if validator.Address == "" {
				validator.Address = change.Address
				validator.Proxy = change.Address
			}
			if change.Proxy != "" {
				if change.Proxy == types.ZeroAddress {
					validator.Proxy = change.Address
				} else {
					validator.Proxy = change.Proxy
				}
			}
			if change.Amount != "0" {
				validator.Amount = BigIntAdd(validator.Amount, change.Amount)
			}
			validator.Timestamp = int64(wh.Timestamp)
			validator.LastNumber = int64(wh.Number)
			err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"proxy", "amount", "timestamp", "last_number"}),
			}).Create(&validator).Error
			if err != nil {
				return
			}
		}
	}
	return
}

func saveExchanger(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, change := range wh.ChangeExchangers {
		exchanger := model.Exchanger{Address: change.Address, Creator: change.Address, Amount: "0", Reward: "0", TxAmount: "0"}
		if err = db.Where("address=?", change.Address).Find(&exchanger).Error; err != nil {
			return
		}
		if change.CloseAt != nil && exchanger.Amount != "0" {
			// close exchanger
			err = db.Select("amount", "close_at").Updates(change).Error
			change.Amount = "-" + exchanger.Amount
		} else {
			// open exchanger
			if change.Creator != "" {
				exchanger.Name = change.Name
				exchanger.URL = change.URL
				exchanger.FeeRatio = change.FeeRatio
				exchanger.Timestamp = change.Timestamp
				exchanger.BlockNumber = change.BlockNumber
				exchanger.TxHash = change.TxHash
				exchanger.CloseAt = nil
			}
			// exchanger pledge
			if change.Amount != "0" {
				exchanger.Amount = BigIntAdd(exchanger.Amount, change.Amount)
			}
			err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"name", "url", "fee_ratio", "timestamp", "block_number", "tx_hash", "amount", "close_at"}),
			}).Create(&exchanger).Error
		}
		if err != nil {
			return
		}
	}
	return
}

func reTryMeta(db *gorm.DB) {
	var NFTs []*model.NFT
	db.Find(&NFTs, "status>0")
	for _, nft := range NFTs {
		nftMeta, err := utils.GetNFTMeta(nft.MetaUrl)
		if err != nil {
			if strings.Index(err.Error(), "context deadline exceeded") > 0 {
				nft.Status += 1
			} else {
				nft.Status = -1
			}
			db.Select("status").Updates(nft)
			log.Println("Failed to retry parse NFT meta information", nft.Address, nft.MetaUrl, err)
			continue
		}

		//collection name + collection creator + hash of the exchange where the collection is located
		if nftMeta.CollectionsName != "" && nftMeta.CollectionsCreator != "" {
			nft.CollectionId = string(utils.Keccak256Hash(
				[]byte(nftMeta.CollectionsName),
				[]byte(nftMeta.CollectionsCreator),
				[]byte(nftMeta.CollectionsExchanger),
			))
			db.Save(&model.Collection{
				Id:          nft.CollectionId,
				Name:        nftMeta.CollectionsName,
				Creator:     nftMeta.CollectionsCreator,
				Category:    nftMeta.CollectionsCategory,
				Desc:        nftMeta.CollectionsDesc,
				ImgUrl:      nftMeta.CollectionsImgUrl,
				BlockNumber: nft.BlockNumber,
				Exchanger:   &nftMeta.CollectionsExchanger,
			})
		}
		nft.Status = 0
		nft.Name = nftMeta.Name
		nft.Desc = nftMeta.Desc
		nft.Attributes = nftMeta.Attributes
		nft.Category = nftMeta.Category
		nft.SourceUrl = nftMeta.SourceUrl
		db.Select("status", "name", "desc", "attributes", "category", "source_url", "collection_id").Updates(nft)
	}
	log.Printf("retry parse nft meta: %v\n", len(NFTs))

	var FNFTs []*model.FNFT
	db.Find(&FNFTs, "status>0")
	for _, fnft := range FNFTs {
		nftMeta, err := utils.GetNFTMeta(fnft.MetaUrl)
		if err != nil {
			if strings.Index(err.Error(), "context deadline exceeded") > 0 {
				fnft.Status += 1
			} else {
				fnft.Status = -1
			}
			db.Select("status").Updates(fnft)
			log.Println("Failed to retry parse SNFT meta information", fnft.ID, fnft.MetaUrl, err)
			continue
		}
		//collection name + collection creator + hash of the exchange where the collection is located
		if fnft.ID[40] == '0' {
			db.Select("name", "creator", "category", "desc", "img_url", "exchanger").Updates(&model.Collection{
				Id:        fnft.ID[:40],
				Name:      nftMeta.CollectionsName,
				Creator:   nftMeta.CollectionsCreator,
				Category:  nftMeta.CollectionsCategory,
				Desc:      nftMeta.CollectionsDesc,
				ImgUrl:    nftMeta.CollectionsImgUrl,
				Exchanger: &nftMeta.CollectionsExchanger,
			})
		}
		fnft.Status = 0
		fnft.Name = nftMeta.Name
		fnft.Desc = nftMeta.Desc
		fnft.Attributes = nftMeta.Attributes
		fnft.Category = nftMeta.Category
		fnft.SourceUrl = nftMeta.SourceUrl
		db.Select("status", "name", "desc", "attributes", "category", "source_url", "collection_id").Updates(fnft)
	}
	log.Printf("retry parse snft meta: %v\n", len(FNFTs))
}
