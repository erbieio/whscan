package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

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

func FixHead(parsed *model.Parsed) (types.Uint64, error) {
	var errBlocks []*model.Block
	if DB.Find(&errBlocks, "number>?", parsed.Number).Error == nil {
		data, _ := json.Marshal(errBlocks)
		log.Printf("err block: %s", string(data))
	}
	return parsed.Number + 1, DB.Transaction(func(tx *gorm.DB) (err error) {
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

func BlockInsert(parsed *model.Parsed) (blocks []types.Hash, err error) {
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
	// exchange creation
	if wh.Exchangers != nil {
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"name", "url", "fee_ratio", "timestamp", "block_number", "tx_hash", "amount", "count", "close_at"}),
		}).Create(wh.Exchangers).Error
		if err != nil {
			return
		}
	}
	// Officially inject SNFT meta information
	for _, epoch := range wh.Epochs {
		err = InjectSNFT(tx, epoch)
		if err != nil {
			return
		}
	}
	// Recycle SNFT
	for _, snftTx := range wh.RecycleTxs {
		var snfts []string
		if err = tx.Model(&model.SNFT{}).Where("LEFT(address,?)=?", len(snftTx.Address), snftTx.Address).Pluck("address", &snfts).Error; err != nil {
			return
		}
		snftTx.Count = int64(len(snfts))
		if err = tx.Create(snftTx).Error; err != nil {
			return
		}
		wh.RecycleSNFTs = append(wh.RecycleSNFTs, snfts...)
		if err = tx.Delete(&model.SNFT{}, "address IN (?)", snfts).Error; err != nil {
			return
		}
	}
	// SNFT reward
	if wh.RewardSNFTs != nil {
		err = tx.Create(wh.RewardSNFTs).Error
		if err != nil {
			return
		}
	}
	// NFT creation
	err = SaveNFT(tx, wh.Number, wh.CreateNFTs)
	if err != nil {
		return
	}
	// NFT transactions, including user and official types of NFTs
	for _, nftTx := range wh.NFTTxs {
		if err = SaveNFTTx(tx, nftTx); err != nil {
			return
		}
	}
	// exchanger pledge
	for _, pledge := range wh.ExchangerPledges {
		var exchanger model.Exchanger
		if err = tx.Where("address=?", pledge.Address).Find(&exchanger).Error; err != nil {
			return
		}
		if exchanger.Address != pledge.Address {
			exchanger.Address = pledge.Address
			exchanger.Amount = "0"
			exchanger.TxAmount = "0"
		}
		exchanger.Count++
		exchanger.Amount = BigIntAdd(exchanger.Amount, pledge.Amount)
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"amount", "count"}),
		}).Create(&exchanger).Error
		if err != nil {
			return
		}
	}
	// close exchanger
	for _, exchanger := range wh.CloseExchangers {
		err = tx.Model(&model.Exchanger{}).Where("address=?", exchanger).Updates(map[string]any{"close_at": wh.Timestamp, "amount": "0"}).Error
		if err != nil {
			return
		}
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
		if change.Proxy != nil {
			if *change.Proxy == "0x0000000000000000000000000000000000000000" || *change.Proxy == change.Address {
				validator.Proxy = nil
			} else {
				validator.Proxy = change.Proxy
			}
		}
		if change.Amount != "0" {
			validator.Amount = BigIntAdd(validator.Amount, change.Amount)
			validator.Count++
		}
		validator.Timestamp = uint64(wh.Timestamp)
		validator.LastNumber = uint64(wh.Number)
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"proxy", "amount", "count", "timestamp", "last_number"}),
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
		if reward.Amount != nil {
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
	return
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
	count, pledgeAmount := db.RowsAffected, "0"
	switch 42 - len(snft) {
	case 0:
		b := big.NewInt(95 * count)
		pledgeAmount = b.Mul(b, big.NewInt(1000000000000000)).Text(10)
	case 1:
		b := big.NewInt(143 * count)
		pledgeAmount = b.Mul(b, big.NewInt(1000000000000000)).Text(10)
	case 2:
		b := big.NewInt(271 * count)
		pledgeAmount = b.Mul(b, big.NewInt(1000000000000000)).Text(10)
	default:
		b := big.NewInt(650 * count)
		pledgeAmount = b.Mul(b, big.NewInt(1000000000000000)).Text(10)
	}
	amount := "0"
	if err = tx.Model(&model.Account{}).Where("address=?", owner).Pluck("snft_amount", &amount).Error; err != nil {
		return
	}
	if isPledge {
		amount = BigIntAdd(amount, pledgeAmount)
	} else {
		amount = BigIntAdd(amount, "-"+pledgeAmount)
	}
	err = tx.Model(&model.Account{}).Where("address=?", owner).Update("snft_amount", amount).Error
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
func SaveNFTTx(tx *gorm.DB, nt *model.NFTTx) error {
	if (*nt.NFTAddr)[2] != '8' {
		// handle NFT
		var nft model.NFT
		err := tx.Where("address=?", nt.NFTAddr).First(&nft).Error
		if err != nil {
			return err
		}
		// populate seller field (if none)
		if nt.From == "" {
			nt.From = nft.Owner
		}
		if nt.Price != nil {
			nft.LastPrice = nt.Price
			nft.TxAmount = BigIntAdd(nft.TxAmount, *nt.Price)
		}
		nft.Owner = nt.To
		err = tx.Select("last_price", "tx_amount", "owner").Updates(&nft).Error
		if err != nil {
			return err
		}
	} else {
		var snfts []*model.SNFT
		err := tx.Where("LEFT(address,?)=?", len(*nt.NFTAddr), *nt.NFTAddr).Find(&snfts).Error
		if err != nil {
			return err
		}
		// populate seller field (if none)
		if nt.From == "" {
			nt.From = snfts[0].Owner
		}
		for _, snft := range snfts {
			if nt.Price != nil {
				snft.TxAmount = BigIntAdd(snft.TxAmount, *nt.Price)
				snft.LastPrice = nt.Price
			}
			snft.Owner = nt.To
			err = tx.Select("last_price", "tx_amount", "owner").Updates(snft).Error
			if err != nil {
				return err
			}
		}
	}
	// Calculate the total number of transactions and the total transaction amount to fill the NFT transaction fee and save the exchange
	if len(nt.ExchangerAddr) == 42 && nt.Price != nil && *nt.Price != "0" {
		var exchanger model.Exchanger
		tx.Find(&exchanger, "address=?", nt.ExchangerAddr)
		if exchanger.Address == nt.ExchangerAddr {
			nt.Fee = TxFee(*nt.Price, exchanger.FeeRatio)
			exchanger.TxAmount = BigIntAdd(exchanger.TxAmount, *nt.Price)
			if err := tx.Select("tx_amount").Updates(&exchanger).Error; err != nil {
				return err
			}
		}
	}
	return tx.Create(&nt).Error
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
