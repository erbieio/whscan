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

// getNFTAddr Get the NFT address
func getNFTAddr(next *big.Int) string {
	return string(utils.BigToAddress(next.Add(next, big.NewInt(stats.TotalNFT+1))))
}

func SetHead(parsed *model.Parsed) error {
	var errBlocks []*model.Block
	if DB.Find(&errBlocks, "number>?", parsed.Number).Error == nil {
		data, _ := json.Marshal(errBlocks)
		log.Printf("err block: %s", string(data))
	}
	return DB.Transaction(func(tx *gorm.DB) (err error) {
		if head := parsed.Number; head != ^types.Uint64(0) {
			if err = tx.Delete(&model.FNFT{}, "LEFT(`id`, 39) IN (?)",
				tx.Model(&model.Epoch{}).Select("id").Where("number>?", head),
			).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Epoch{}, "number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Collection{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Exchanger{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Model(&model.Exchanger{}).Where("close_at>?", head).Update("close_at", nil).Error; err != nil {
				return
			}
			hashes := tx.Model(&model.Transaction{}).Select("hash").Where("block_number>?", head)
			if err = tx.Delete(&model.Account{}, "created_tx IN (?)", hashes).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.ERC20Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.ERC721Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.ERC1155Transfer{}, "tx_hash IN (?)", hashes).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.InternalTx{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Log{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.NFTTx{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.NFT{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.SNFT{}, "reward_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Reward{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Transaction{}, "block_number>?", head).Error; err != nil {
				return
			}
			if err = tx.Delete(&model.Block{}, "number>?", head).Error; err != nil {
				return
			}
		} else {
			if err = model.ClearTable(tx); err != nil {
				return
			}
		}
		return fixStats(tx, parsed)
	})
}

func Insert(parsed *model.Parsed) (blocks []types.Hash, err error) {
	if parsed.Number > 0 {
		err = DB.Take(&model.Block{}, "number=? AND hash=?", parsed.Number-1, parsed.ParentHash).Error
		if err == gorm.ErrRecordNotFound {
			err = DB.Model(&model.Block{}).Order("number DESC").Limit(1000).Pluck("hash", &blocks).Error
			return
		}
	}
	err = DB.Transaction(func(db *gorm.DB) error {
		for _, tx := range parsed.CacheTxs {
			// write block transaction
			if err := db.Create(tx).Error; err != nil {
				return err
			}
		}

		for _, internalTx := range parsed.CacheInternalTxs {
			// write internal transaction
			if err := db.Create(internalTx).Error; err != nil {
				return err
			}
		}

		for _, uncle := range parsed.CacheUncles {
			// write uncle block
			if err := db.Create(uncle).Error; err != nil {
				return err
			}
		}

		for _, account := range parsed.CacheAccounts {
			// write account information
			if err := db.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"balance", "nonce"}),
			}).Create(account).Error; err != nil {
				return err
			}
		}

		// write transaction cacheLog
		for _, cacheLog := range parsed.CacheLogs {
			if err := db.Create(cacheLog).Error; err != nil {
				return err
			}
		}
		for _, cacheTransferLog := range parsed.CacheTransferLogs {
			if err := db.Create(cacheTransferLog).Error; err != nil {
				return err
			}
		}

		// write block
		if err := db.Create(parsed.Block).Error; err != nil {
			return err
		}

		// wormholes unique data write
		if err := WHInsert(db, parsed); err != nil {
			return err
		}

		// update the query stats
		return updateStats(db, parsed)
	})
	freshStats(DB)
	return
}

func WHInsert(tx *gorm.DB, wh *model.Parsed) (err error) {
	// Officially inject SNFT meta information
	if wh.Epoch != nil {
		if err = InjectSNFT(tx, wh.Epoch); err != nil {
			return
		}
	}
	// NFT creation
	err = SaveNFT(tx, wh.Number, wh.NFTs)
	if err != nil {
		return
	}
	// NFT transactions, including user and official types of NFTs
	if err = SaveNFTTx(tx, wh); err != nil {
		return
	}
	// validator change
	for _, change := range wh.ChangeValidators {
		var validator model.Validator
		if err = tx.Find(&validator, "address=?", change.Address).Error; err != nil {
			return
		}
		if change.Address != validator.Address {
			validator.Address = change.Address
			validator.Amount = "0"
			validator.Reward = "0"
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
		validator.Timestamp = uint64(wh.Timestamp)
		validator.LastNumber = uint64(wh.Number)
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"proxy", "amount", "timestamp", "last_number"}),
		}).Create(&validator).Error
		if err != nil {
			return
		}
	}
	for _, snft := range wh.PledgeSNFT {
		if err = SaveSNFTPledge(tx, snft[:42], snft[42:], wh.Number, true); err != nil {
			return
		}
	}
	for _, snft := range wh.UnPledgeSNFT {
		if err = SaveSNFTPledge(tx, snft[:42], snft[42:], wh.Number, false); err != nil {
			return
		}
	}
	for _, reward := range wh.Rewards {
		err = tx.Exec("INSERT INTO rewards (`address`, `proxy`, `identity`, `block_number`, `snft`, `amount`) VALUES "+
			"(@Address, (SELECT `proxy` FROM validators WHERE address=@Address), @Identity, @BlockNumber, @SNFT, @Amount)", reward).Error
		if err != nil {
			return
		}
		if reward.SNFT != nil {
			err = tx.Create(&model.SNFT{
				Address:      *reward.SNFT,
				TxAmount:     "0",
				Awardee:      reward.Address,
				RewardAt:     uint64(wh.Timestamp),
				RewardNumber: uint64(wh.Number),
				Owner:        reward.Address,
			}).Error
			if err != nil {
				return
			}
			var user *model.User
			var exchanger *model.Exchanger
			if err = tx.Model(&model.User{}).Where("`amount`!='0' AND `address`=?", reward.Address).Scan(&user).Error; err != nil {
				return
			}
			if err = tx.Model(&model.Exchanger{}).Where("`amount`!='0' AND `address`=?", reward.Address).Scan(&exchanger).Error; err != nil {
				return
			}
			user, exchanger = updateReward(big.NewInt(95000000000000000), user, exchanger)
			if user != nil {
				if err = tx.Updates(user).Error; err != nil {
					return
				}
			}
			if exchanger != nil {
				if err = tx.Updates(exchanger).Error; err != nil {
					return
				}
			}
		} else {
			var pledge model.Validator
			if err = tx.Find(&pledge, "address=?", reward.Address).Error; err != nil {
				return
			}
			if len(pledge.Address) != 0 {
				pledge.Reward = BigIntAdd(pledge.Reward, *reward.Amount)
				pledge.Timestamp = uint64(wh.Timestamp)
				pledge.LastNumber = uint64(wh.Number)
				if err = tx.Updates(&pledge).Error; err != nil {
					return
				}
			}
		}
	}
	for _, change := range wh.ChangeExchangers {
		exchanger := model.Exchanger{Address: change.Address, Creator: change.Address, Amount: "0", Reward: "0", TxAmount: "0"}
		if err = tx.Where("address=?", change.Address).Find(&exchanger).Error; err != nil {
			return
		}
		if change.CloseAt != nil && exchanger.Amount != "0" {
			// close exchanger
			err = tx.Select("amount", "close_at").Updates(change).Error
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
			}
			// exchanger pledge
			if change.Amount != "0" {
				exchanger.Amount = BigIntAdd(exchanger.Amount, change.Amount)
			}
			err = tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"name", "url", "fee_ratio", "timestamp", "block_number", "tx_hash", "amount"}),
			}).Create(&exchanger).Error
		}
		if err != nil {
			return
		}
	}
	return
}

func snftValue(snft string, count int64) string {
	b := big.NewInt(count)
	switch 42 - len(snft) {
	case 0:
		return b.Mul(b, big.NewInt(95000000000000000)).Text(10)
	case 1:
		return b.Mul(b, big.NewInt(143000000000000000)).Text(10)
	case 2:
		return b.Mul(b, big.NewInt(271000000000000000)).Text(10)
	default:
		return b.Mul(b, big.NewInt(650000000000000000)).Text(10)
	}
}

func SaveSNFTPledge(tx *gorm.DB, owner, snft string, number types.Uint64, isPledge bool) (err error) {
	var db *gorm.DB
	if isPledge {
		db = tx.Model(&model.SNFT{}).Where("LEFT(address,?)=?", len(snft), snft).Update("pledge_number", number)
	} else {
		db = tx.Model(&model.SNFT{}).Where("LEFT(address,?)=?", len(snft), snft).Update("pledge_number", nil)
	}
	if err = db.Error; err != nil {
		return
	}
	amount, pledgeAmount := "0", snftValue(snft, db.RowsAffected)
	if err = tx.Model(&model.User{}).Where("address=?", owner).Pluck("amount", &amount).Error; err != nil {
		return
	}
	if isPledge {
		amount = BigIntAdd(amount, pledgeAmount)
	} else {
		amount = BigIntAdd(amount, "-"+pledgeAmount)
	}
	if err = tx.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"amount"}),
	}).Create(&model.User{Address: owner, Amount: amount, Reward: "0"}).Error; err != nil {
		return
	}
	return
}

// SaveNFT saves the NFT created by the user
func SaveNFT(tx *gorm.DB, number types.Uint64, nfts []*model.NFT) error {
	for i, nft := range nfts {
		*nft.Address = getNFTAddr(big.NewInt(int64(i)))
		err := tx.Create(nft).Error
		if err != nil {
			return err
		}
		// Update the total number of NFTs on the specified exchange
		if len(nft.ExchangerAddr) == 42 {
			var exchanger model.Exchanger
			tx.Find(&exchanger, "address=?", nft.ExchangerAddr)
			if exchanger.Address == nft.ExchangerAddr {
				exchanger.NFTCount++
				if err = tx.Select("nft_count").Updates(&exchanger).Error; err != nil {
					return err
				}
			}
		}
		if nft.MetaUrl != "" {
			go saveNFTMeta(number, *nft.Address, nft.MetaUrl)
		}
	}
	return nil
}

// SaveNFTTx saves the NFT transaction record, while updating the NFT owner and the latest price
func SaveNFTTx(db *gorm.DB, wh *model.Parsed) (err error) {
	for _, tx := range wh.NFTTxs {
		if tx.To == "" {
			// handle recycle snft
			var snfts []string
			if err = db.Model(&model.SNFT{}).Where("LEFT(address,?)=?", len(*tx.NFTAddr), *tx.NFTAddr).Pluck("address", &snfts).Error; err != nil {
				return
			}
			fee := strconv.Itoa(len(snfts))
			tx.Fee = &fee
			tx.Price = snftValue(*tx.NFTAddr, int64(len(snfts)))
			wh.RecycleSNFTs = append(wh.RecycleSNFTs, snfts...)
			if err = db.Delete(&model.SNFT{}, "address IN (?)", snfts).Error; err != nil {
				return
			}
		} else {
			if (*tx.NFTAddr)[2] != '8' {
				// handle NFT
				var nft model.NFT
				if err = db.Where("address=?", tx.NFTAddr).First(&nft).Error; err != nil {
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
				var snfts []*model.SNFT
				if err = db.Where("LEFT(address,?)=?", len(*tx.NFTAddr), *tx.NFTAddr).Find(&snfts).Error; err != nil {
					return err
				}
				// populate seller field (if none)
				if tx.From == "" {
					tx.From = snfts[0].Owner
				}
				for _, snft := range snfts {
					if tx.Price != "0" {
						snft.LastPrice = &tx.Price
						snft.TxAmount = BigIntAdd(snft.TxAmount, tx.Price)
					}
					snft.Owner = tx.To
					if err = db.Select("last_price", "tx_amount", "owner").Updates(snft).Error; err != nil {
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

// InjectSNFT official batch injection of SNFT
func InjectSNFT(tx *gorm.DB, epoch *model.Epoch) (err error) {
	err = tx.Create(epoch).Error
	if err != nil {
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
		err = tx.
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
			err = tx.Create(&model.FNFT{ID: FNFTId, MetaUrl: metaUrl}).Error
			if err != nil {
				return
			}
			if metaUrl != "" {
				go saveSNFTMeta(FNFTId, metaUrl)
			}
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
func saveNFTMeta(blockNumber types.Uint64, nftAddr, metaUrl string) {
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
			BlockNumber: uint64(blockNumber),
			Exchanger:   &nftMeta.CollectionsExchanger,
		})
		err = result.Error
	}
}
