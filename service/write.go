package service

import (
	"encoding/json"
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

		// erbie unique data write
		// NFT transactions, including user and official types of NFTs
		if err = saveErbie(db, parsed); err != nil {
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
		if err = saveSlashing(db, parsed); err != nil {
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
			if err = db.Delete(&model.Epoch{}, "start_number>?", head).Error; err != nil {
				return
			}
			if err = db.Delete(&model.Staker{}, "block_number>?", head).Error; err != nil {
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
			if err = db.Delete(&model.Erbie{}, "block_number>?", head).Error; err != nil {
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
		err = db.Model(&model.Block{}).Where("number=?", epoch.Number).Select("timestamp").Scan(&epoch.Timestamp).Error
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
	if len(wh.Rewards) > 0 {
		if err = db.Create(wh.Rewards).Error; err != nil {
			return
		}
		if err = db.Exec("UPDATE rewards SET proxy=(SELECT proxy FROM validators WHERE address=rewards.address) WHERE block_number=? AND snft=''", wh.Number).Error; err != nil {
			return
		}
		if err = db.Exec("UPDATE validators SET weight=70 WHERE address IN (SELECT address FROM rewards WHERE block_number=? AND snft='')", wh.Number).Error; err != nil {
			return
		}
	} else if len(wh.Proposers) > 0 {
		if err = db.Model(&model.Validator{}).Where("address IN (?)", wh.Proposers).Update("weight", 70).Error; err != nil {
			return
		}
	}

	for _, reward := range wh.Rewards {
		if reward.SNFT != "" {
			err = db.Create(&model.SNFT{
				Address:      reward.SNFT,
				TxAmount:     "0",
				RewardAt:     int64(wh.Timestamp),
				RewardNumber: int64(wh.Number),
				Owner:        reward.Address,
				Pieces:       1,
			}).Error
			if err != nil {
				return
			}
			var staker model.Staker
			if err = db.Find(&staker, "`address`=?", reward.Address).Error; err != nil {
				return
			}
			if staker.Address != "" {
				staker.Reward = BigIntAdd(staker.Reward, snftValue(reward.SNFT, 1))
				staker.RewardCount++
				if err = db.Select("reward", "reward_count").Updates(&staker).Error; err != nil {
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
				validator.RewardNumber = int64(wh.Number)
				validator.Timestamp = int64(wh.Timestamp)
				validator.BlockNumber = int64(wh.Number)
				if err = db.Select("reward", "reward_count", "reward_number", "timestamp", "block_number").Updates(&validator).Error; err != nil {
					return
				}
			}
		}
	}
	return
}

func saveSlashing(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, slashing := range wh.Slashings {
		if err = db.Create(slashing).Error; err != nil {
			return
		}
		if slashing.Reason == "1" {
			if err = db.Model(&model.Validator{}).Where("address=?", slashing.Address).Update("weight", slashing.Weight).Error; err != nil {
				return
			}
		} else {
			err = db.Exec("UPDATE slashings SET amount=(SELECT amount FROM validators WHERE address=slashings.address) WHERE address=? AND block_number=?", slashing.Address, slashing.BlockNumber).Error
			if err != nil {
				return
			}
			err = db.Exec("DELETE FROM validators WHERE address=?", slashing.Address).Error
			if err != nil {
				return
			}
			err = db.Exec("INSERT INTO slashings (SELECT staker AS address, ? AS block_number, amount, 0 AS weight, validator AS reason FROM pledges WHERE validator=?)", slashing.BlockNumber, slashing.Address).Error
			if err != nil {
				return
			}
			err = db.Exec("UPDATE stakers SET amount=amount-(SELECT amount FROM pledges WHERE staker=address AND validator=?) WHERE address IN (SELECT staker FROM pledges WHERE validator=?)", slashing.Address, slashing.Address).Error
			if err != nil {
				return
			}
			err = db.Exec("DELETE FROM pledges WHERE validator=?", slashing.Address).Error
			if err != nil {
				return
			}
		}
	}
	if err = db.Where("amount=0").Delete(&model.Staker{}).Error; err != nil {
		return
	}
	return
}

// saveErbie saves the NFT transaction record, while updating the NFT owner and the latest price
func saveErbie(db *gorm.DB, wh *model.Parsed) (err error) {
	nftCount := int64(0)
	for _, erbie := range wh.Erbies {
		switch erbie.Type {
		case 0: //mint NFT
			nftCount++
			erbie.Address = string(utils.BigToAddress(big.NewInt(nftCount + stats.TotalNFT)))
			if err = db.Create(&model.NFT{
				Address:      erbie.Address,
				RoyaltyRatio: erbie.RoyaltyRate,
				MetaUrl:      erbie.Extra,
				TxAmount:     "0",
				Creator:      erbie.To,
				Timestamp:    erbie.Timestamp,
				BlockNumber:  erbie.BlockNumber,
				TxHash:       erbie.TxHash,
				Owner:        erbie.To,
			}).Error; err != nil {
				return
			}
			db.Exec("UPDATE accounts SET nft_count=nft_count+1 WHERE address=?", erbie.To)

		case 1: //transfer NFT or SNFT
			if erbie.Address[2] == '0' {
				if err = db.Model(&model.NFT{Address: erbie.Address}).Update("owner", erbie.To).Error; err != nil {
					return
				}
				db.Exec("UPDATE accounts SET nft_count=nft_count+1 WHERE address=?", erbie.To)
				db.Exec("UPDATE accounts SET nft_count=nft_count-1 WHERE address=?", erbie.From)
			} else {
				if err = db.Model(&model.SNFT{Address: erbie.Address}).Update("owner", erbie.To).Error; err != nil {
					return
				}
			}

		case 6: //recycle SNFT
			snft := model.SNFT{}
			if err = db.Where("address=?", erbie.Address).Take(&snft).Error; err != nil {
				return
			}
			erbie.FeeRate = snft.Pieces
			erbie.Value = snftValue(erbie.Address, snft.Pieces)
			if err = db.Model(&snft).Update("remove", true).Error; err != nil {
				return
			}

		case 9, 10: //staker pledge to validator
			from, to, value := erbie.From, erbie.To, erbie.Value
			if erbie.Type == 10 {
				value = "-" + value
			}
			pledge := model.Pledge{Staker: from, Validator: to, Amount: "0", TxHash: erbie.TxHash}
			if err = db.Find(&pledge).Error; err != nil {
				return
			}
			pledge.Amount = BigIntAdd(pledge.Amount, value)
			pledge.Timestamp = erbie.Timestamp
			pledge.BlockNumber = erbie.BlockNumber
			if err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"amount", "timestamp", "block_number"}),
			}).Create(&pledge).Error; err != nil {
				return
			}

			staker := model.Staker{Address: from, Amount: "0", Reward: "0", TxHash: erbie.TxHash}
			if err = db.Find(&staker).Error; err != nil {
				return
			}
			staker.FeeRate = erbie.FeeRate
			staker.Amount = BigIntAdd(staker.Amount, value)
			staker.Timestamp = erbie.Timestamp
			staker.BlockNumber = erbie.BlockNumber
			if err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"fee_rate", "amount", "timestamp", "block_number"}),
			}).Create(&staker).Error; err != nil {
				return
			}

			validator := model.Validator{Address: to, Proxy: to, Amount: "0", Reward: "0", TxHash: erbie.TxHash, Weight: 70}
			if err = db.Find(&validator).Error; err != nil {
				return
			}
			validator.Amount = BigIntAdd(validator.Amount, value)
			validator.Timestamp = erbie.Timestamp
			validator.BlockNumber = erbie.BlockNumber
			if err = db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"amount", "timestamp", "block_number"}),
			}).Create(&validator).Error; err != nil {
				return
			}

			if erbie.Type == 10 {
				db.Where("amount=0").Delete(&pledge)
				db.Where("amount=0").Delete(&staker)
				db.Where("amount=0").Delete(&validator)
			}

		case 14, 15, 18, 20, 27:
			//14: trade NFT or SNFT, exchanger or seller send tx
			//15: trade NFT or SNFT, buyer send tx
			//18: trade NFT or SNFT, exchanger proxy send tx
			//20: trade NFT or SNFT, exchanger send tx
			//27: trade NFT or SNFT, exchanger send tx
			if err = db.Model(&model.Staker{Address: erbie.Extra}).Select("IFNULL(fee_rate,0)").Scan(&erbie.FeeRate).Error; err != nil {
				return
			}
			if erbie.Address[2] == '0' {
				if err = db.Model(&model.NFT{Address: erbie.Address}).Updates(map[string]any{
					"last_price": erbie.Value,
					"tx_amount":  gorm.Expr("tx_amount+?", erbie.Value),
					"owner":      erbie.To,
				}).Error; err != nil {
					return
				}
				if err = db.Model(&model.NFT{Address: erbie.Address}).Select("royalty_ratio").Scan(&erbie.RoyaltyRate).Error; err != nil {
					return
				}
				db.Exec("UPDATE accounts SET nft_count=nft_count+1 WHERE address=?", erbie.From)
				db.Exec("UPDATE accounts SET nft_count=nft_count-1 WHERE address=?", erbie.To)
			} else {
				if err = db.Model(&model.SNFT{Address: erbie.Address}).Updates(map[string]any{
					"last_price": erbie.Value,
					"tx_amount":  gorm.Expr("tx_amount+?", erbie.Value),
					"owner":      erbie.To,
				}).Error; err != nil {
					return
				}
				erbie.RoyaltyRate = 1000
				// creator and epoch profit stats
				if erbie.Value != "0" {
					royalty := TxFee(erbie.Value, erbie.RoyaltyRate)
					if err = db.Model(&model.Epoch{ID: (erbie.Address)[:39]}).Update("profit", gorm.Expr("profit+?", royalty)).Error; err != nil {
						return
					}
					if err = db.Model(&model.Creator{}).
						Where("address=(?)", db.Model(&model.Epoch{ID: (erbie.Address)[:39]}).Select("creator")).
						Update("profit", gorm.Expr("profit+?", royalty)).Error; err != nil {
						return
					}
				}
			}

		case 16, 17, 19:
			//16: trade NFT(no minted), buyer send tx
			//17: trade NFT(no minted), exchanger send tx
			//19: trade NFT(no minted), exchanger proxy send tx
			//warn: not process exchanger fee rate
			nftCount++
			erbie.Address = string(utils.BigToAddress(big.NewInt(nftCount + stats.TotalNFT)))
			if err = db.Create(&model.NFT{
				Address:      erbie.Address,
				RoyaltyRatio: erbie.RoyaltyRate,
				MetaUrl:      erbie.Extra,
				LastPrice:    &erbie.Value,
				TxAmount:     erbie.Value,
				Creator:      erbie.From,
				Timestamp:    erbie.Timestamp,
				BlockNumber:  erbie.BlockNumber,
				TxHash:       erbie.TxHash,
				Owner:        erbie.To,
			}).Error; err != nil {
				return
			}
			db.Exec("UPDATE accounts SET nft_count=nft_count+1 WHERE address=?", erbie.To)

		case 26: //recover validator online weight
			if err = db.Updates(&model.Validator{
				Address:     erbie.From,
				Timestamp:   erbie.Timestamp,
				BlockNumber: erbie.BlockNumber,
				Weight:      70,
			}).Error; err != nil {
				return
			}

		case 28: //forcibly buy snft
			addr := [16]string{}
			for i := int64(0); i < 16; i++ {
				addr[i] = erbie.Address + strconv.FormatInt(i, 16)
			}
			result := db.Model(&model.SNFT{}).Where("address IN (?) AND owner!=?", addr, erbie.To).Update("owner", erbie.To)
			if err = result.Error; err != nil {
				return
			}
			erbie.Value = snftValue(erbie.Address, result.RowsAffected)
			erbie.RoyaltyRate = 1000

		case 31: //validator set proxy
			if erbie.To == types.ZeroAddress {
				erbie.To = erbie.From
			}
			if err = db.Updates(&model.Validator{
				Address:     erbie.From,
				Proxy:       erbie.To,
				Timestamp:   erbie.Timestamp,
				BlockNumber: erbie.BlockNumber,
			}).Error; err != nil {
				return
			}
		}
		if erbie.TxHash != "0x0" {
			if err = db.Create(&erbie).Error; err != nil {
				return
			}
		}
	}
	return
}
