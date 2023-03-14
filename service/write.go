package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"

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
		for _, account := range parsed.CacheAccounts {
			err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"balance", "nonce"})}).Create(account).Error
			if err != nil {
				return
			}
		}
		// write block
		if err = db.Create(parsed.Block).Error; err != nil {
			return
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
		if err = savePledge(db, parsed); err != nil {
			return
		}

		// update the query stats
		return updateStats(db, parsed)
	})
	freshStats(DB, parsed)
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
		err = db.Model(&Transaction{}).Where("block_number>? AND block_number<=? AND status=1", epoch.Number-1024, epoch.Number).
			Where("LEFT(input,76)='0x776f726d686f6c65733a7b2276657273696f6e223a22302e3031222c2274797065223a3233'").
			Or("LEFT(input,76)='0x776f726d686f6c65733a7b2276657273696f6e223a22302e3031222c2274797065223a3234'").
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
		for i := 0; i < 16; i++ {
			hexI := fmt.Sprintf("%x", i)
			collectionId := epoch.ID + hexI
			metaUrl := ""
			if epoch.Dir != "" {
				metaUrl = epoch.Dir + hexI + "0"
			}
			// write collection information
			err = db.
				Create(&model.Collection{Id: collectionId, MetaUrl: metaUrl, BlockNumber: epoch.Number}).Error
			if err != nil {
				return
			}
			if metaUrl != "" {
				go saveSNFTCollection(collectionId, metaUrl)
			}
			for j := 0; j < 16; j++ {
				hexJ := fmt.Sprintf("%x", j)
				FNFTId := collectionId + hexJ
				if epoch.Dir != "" {
					metaUrl = epoch.Dir + hexI + hexJ
				}
				// write complete SNFT information
				err = db.Create(&model.FNFT{ID: FNFTId, MetaUrl: metaUrl}).Error
				if err != nil {
					return
				}
				if metaUrl != "" {
					go saveSNFTMeta(FNFTId, metaUrl)
				}
			}
		}
	}
	return
}

func updateUserSNFT(db *gorm.DB, number types.Long, user, value string, count int64) (err error) {
	var account model.Account
	if err = db.Find(&account, "`address`=?", user).Error; err != nil {
		return
	}
	if account.Address == "" {
		account.Address = types.Address(user)
		account.Balance = "0"
		account.Number = number
		account.SNFTValue = value
		account.SNFTCount = count
		err = db.Create(&account).Error
	} else {
		account.Number = number
		account.SNFTValue = BigIntAdd(account.SNFTValue, value)
		account.SNFTCount += count
		err = db.Select("snft_count", "snft_value", "number").Updates(&account).Error
	}
	return
}

func saveMerge(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, snft := range wh.Mergers {
		result := db.Model(&model.SNFT{}).Where("LEFT(`address`,?)=? AND remove=false", len(snft.Address), snft.Address).Update("remove", true)
		if err = result.Error; err != nil {
			return
		}
		if err = db.Create(snft).Error; err != nil {
			return
		}
		if err = updateUserSNFT(db, wh.Number, snft.Owner, snftMergeValue(snft.Address, snft.Pieces), 1-result.RowsAffected); err != nil {
			return
		}
	}
	return
}

func savePledge(db *gorm.DB, wh *model.Parsed) (err error) {
	if wh.Number == 1 {
		err = db.Model(&model.Pledge{}).Where("timestamp=0").Update("timestamp", wh.Timestamp).Error
		if err != nil {
			return
		}
	}
	if len(wh.Pledges) > 0 {
		err = db.Create(wh.Pledges).Error
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
			if err = updateUserSNFT(db, wh.Number, reward.Address, value, 1); err != nil {
				return
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
			var exchanger model.Exchanger
			db.Find(&exchanger, "address=?", nft.ExchangerAddr)
			if exchanger.Address == nft.ExchangerAddr {
				exchanger.NFTCount++
				if err = db.Select("nft_count").Updates(&exchanger).Error; err != nil {
					return
				}
			}
		}
		if nft.MetaUrl != "" {
			go saveNFTMeta(wh.Number, *nft.Address, nft.MetaUrl)
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
			if err = updateUserSNFT(db, wh.Number, snft.Owner, "-"+tx.Price, -1); err != nil {
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
				if tx.Price != "0" {
					nft.LastPrice = &tx.Price
					nft.TxAmount = BigIntAdd(nft.TxAmount, tx.Price)
				}
				nft.Owner = tx.To
				if err = db.Select("last_price", "tx_amount", "owner").Updates(&nft).Error; err != nil {
					return
				}
			} else {
				var snft model.SNFT
				if err = db.Where("address=?", tx.NFTAddr).Take(&snft).Error; err != nil {
					return
				}
				// populate seller field (if none)
				if tx.From == "" {
					tx.From = snft.Owner
				}
				if tx.Price != "0" {
					snft.LastPrice = &tx.Price
					snft.TxAmount = BigIntAdd(snft.TxAmount, tx.Price)
					creator := model.Creator{}
					err = db.Find(&creator, "address=(?)", db.Model(&model.Epoch{}).Where("`id`=?", snft.Address[:39]).Select("creator")).Error
					if err != nil {
						return
					}
					creator.Profit = BigIntAdd(creator.Profit, *TxFee(tx.Price, 1000))
					if err = db.Select("profit").Updates(creator).Error; err != nil {
						return
					}
				}
				snft.Owner = tx.To
				if err = db.Select("last_price", "tx_amount", "owner").Updates(&snft).Error; err != nil {
					return
				}
				if tx.From != tx.To {
					value := snftValue(snft.Address, snft.Pieces)
					if err = updateUserSNFT(db, wh.Number, tx.From, "-"+value, -1); err != nil {
						return
					}
					if err = updateUserSNFT(db, wh.Number, tx.To, value, 1); err != nil {
						return
					}
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
				if change.Proxy == "0x0000000000000000000000000000000000000000" {
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

func saveSNFTCollection(id, metaUrl string) {
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}
	err = DB.Model(&model.Collection{}).Where("id=?", id).Updates(map[string]any{
		"name":     nftMeta.CollectionsName,
		"desc":     nftMeta.CollectionsDesc,
		"category": nftMeta.CollectionsCategory,
		"img_url":  nftMeta.CollectionsImgUrl,
		"creator":  nftMeta.CollectionsCreator,
	}).Error
	if err != nil {
		log.Println("Failed to parse and store SNFT collection information", id, metaUrl, err)
	}
}

func saveSNFTMeta(id, metaUrl string) {
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}
	err = DB.Model(&model.FNFT{}).Where("id=?", id).Updates(map[string]any{
		"name":       nftMeta.Name,
		"desc":       nftMeta.Desc,
		"attributes": nftMeta.Attributes,
		"category":   nftMeta.Category,
		"source_url": nftMeta.SourceUrl,
	}).Error
	if err != nil {
		log.Println("Failed to parse and store SNFT meta information", id, metaUrl, err)
	}
}

// saveNFTMeta parses and stores NFT meta information
func saveNFTMeta(blockNumber types.Long, nftAddr, metaUrl string) {
	var err error
	defer func() {
		if err != nil {
			log.Println("Failed to parse and store NFT meta information", nftAddr, metaUrl, err)
		}
	}()
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}

	//collection name + collection creator + hash of the exchange where the collection is located
	var collectionId *string
	if nftMeta.CollectionsName != "" && nftMeta.CollectionsCreator != "" {
		hash := string(utils.Keccak256Hash(
			[]byte(nftMeta.CollectionsName),
			[]byte(nftMeta.CollectionsCreator),
			[]byte(nftMeta.CollectionsExchanger),
		))
		collectionId = &hash
	}
	err = DB.Model(&model.NFT{}).Where("address=?", nftAddr).Updates(map[string]any{
		"name":          nftMeta.Name,
		"desc":          nftMeta.Desc,
		"attributes":    nftMeta.Attributes,
		"category":      nftMeta.Category,
		"source_url":    nftMeta.SourceUrl,
		"collection_id": collectionId,
	}).Error
	if err == nil && collectionId != nil {
		result := DB.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&model.Collection{
			Id:          *collectionId,
			Name:        nftMeta.CollectionsName,
			Creator:     nftMeta.CollectionsCreator,
			Category:    nftMeta.CollectionsCategory,
			Desc:        nftMeta.CollectionsDesc,
			ImgUrl:      nftMeta.CollectionsImgUrl,
			BlockNumber: int64(blockNumber),
			Exchanger:   &nftMeta.CollectionsExchanger,
		})
		err = result.Error
	}
}
