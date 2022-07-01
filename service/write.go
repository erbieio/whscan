package service

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// Cache caches some database queries to speed up queries
type Cache struct {
	TotalBlock          uint64       `json:"totalBlock"`          //Total number of blocks
	TotalTransaction    uint64       `json:"totalTransaction"`    //Total number of transactions
	TotalUncle          uint64       `json:"totalUncle"`          //Number of total uncle blocks
	TotalAccount        uint64       `json:"totalAccount"`        //Total account number
	GenesisBalance      types.BigInt `json:"genesisBalance"`      //Total amount of coins created
	TotalBalance        types.BigInt `json:"totalBalance"`        //The total amount of coins in the chain
	TotalExchanger      uint64       `json:"totalExchanger"`      //Total number of exchanges
	TotalUNFTCollection uint64       `json:"totalUNFTCollection"` //Total number of UNFT collections
	TotalSNFTCollection uint64       `json:"totalSNFTCollection"` //Total number of SNFT collections
	TotalUNFT           uint64       `json:"totalUNFT"`           //Total number of UNFTs
	TotalSNFT           uint64       `json:"totalSNFT"`           //Total number of SNFTs
	TotalNFTTx          uint64       `json:"totalNFTTx"`          //Total number of NFT transactions
}

var cache = Cache{}

// InitCache initializes the query cache from the database
func initCache() (err error) {
	if err = DB.Model(&model.Block{}).Select("COUNT(*)").Scan(&cache.TotalBlock).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Transaction{}).Select("COUNT(*)").Scan(&cache.TotalTransaction).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Uncle{}).Select("COUNT(*)").Scan(&cache.TotalUncle).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Account{}).Select("COUNT(*)").Scan(&cache.TotalAccount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "GenesisBalance").Select("value").Scan(&cache.GenesisBalance).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalBalance").Select("value").Scan(&cache.TotalBalance).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Exchanger{}).Select("COUNT(*)").Scan(&cache.TotalExchanger).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Collection{}).Where("length(id)!=39").Select("COUNT(*)").Scan(&cache.TotalUNFTCollection).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Collection{}).Where("length(id)=39").Select("COUNT(*)").Scan(&cache.TotalSNFTCollection).Error; err != nil {
		return
	}
	if err = DB.Model(&model.UNFT{}).Select("COUNT(*)").Scan(&cache.TotalUNFT).Error; err != nil {
		return
	}
	if err = DB.Model(&model.SNFT{}).Select("COUNT(*)").Scan(&cache.TotalSNFT).Error; err != nil {
		return
	}
	if err = DB.Model(&model.NFTTx{}).Select("COUNT(*)").Scan(&cache.TotalNFTTx).Error; err != nil {
		return
	}
	return err
}

func TotalBlock() uint64 {
	return cache.TotalBlock
}

func TotalTransaction() uint64 {
	return cache.TotalTransaction
}

func TotalSNFT() uint64 {
	return cache.TotalSNFT
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
	if cache.TotalBalance == "" {
		cache.TotalBalance = "0"
	}
	return cache.TotalBalance
}

// getNFTAddr Get the NFT address
func getNFTAddr(next *big.Int) string {
	return string(utils.BigToAddress(next.Add(next, big.NewInt(int64(cache.TotalUNFT+1)))))
}

// DecodeRet block parsing result
type DecodeRet struct {
	*model.Block
	CacheTxs         []*model.Transaction `json:"transactions"`
	CacheInternalTxs []*model.InternalTx
	CacheUncles      []*model.Uncle
	CacheAccounts    map[types.Address]*model.Account
	CacheLogs        []*model.Log //Insert after CacheAccounts
	AddBalance       *big.Int     //The number of coins added by the block

	// wormholes, which need to be inserted into the database by priority (later data may query previous data)
	Exchangers       []*model.Exchanger       //The created exchange, priority: 1
	Epochs           []*model.Epoch           //Official injection of the first phase of SNFT, priority: 1
	CreateNFTs       []*model.UNFT            //Newly created NFT, priority: 2
	RewardSNFTs      []*model.SNFT            //Reward information of SNFT, priority: 3
	NFTTxs           []*model.NFTTx           //NFT transaction record, priority: 4
	RecycleSNFTs     []string                 //Recycle SNFT, priority: 5
	CloseExchangers  []string                 //Close exchanges, priority: 5
	Rewards          []*model.Reward          //reward record, priority: none
	ExchangerPledges []*model.ExchangerPledge //Exchange pledge, priority: none
	ConsensusPledges []*model.ConsensusPledge //Consensus pledge, priority: none
}

func BlockInsert(block *DecodeRet) error {
	err := DB.Transaction(func(t *gorm.DB) error {
		for _, tx := range block.CacheTxs {
			// write block transaction
			if err := t.Create(tx).Error; err != nil {
				return err
			}
		}

		for _, internalTx := range block.CacheInternalTxs {
			// write internal transaction
			if err := t.Create(internalTx).Error; err != nil {
				return err
			}
		}

		for _, uncle := range block.CacheUncles {
			// write uncle block
			if err := t.Create(uncle).Error; err != nil {
				return err
			}
		}

		for _, account := range block.CacheAccounts {
			// write account information
			if err := t.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"balance", "nonce"}),
			}).Create(account).Error; err != nil {
				return err
			}
		}

		// write transaction log
		err := SaveTxLog(t, block.CacheLogs)
		if err != nil {
			return err
		}

		// write block
		if err := t.Create(block.Block).Error; err != nil {
			return err
		}

		if block.AddBalance.Uint64() != 0 {
			if block.Number == 0 {
				err = t.Create(model.Cache{
					Key:   "GenesisBalance",
					Value: block.AddBalance.Text(10),
				}).Error
				if err != nil {
					return err
				}
			}
			b := new(big.Int)
			b.SetString(string(cache.TotalBalance), 10)
			b.Add(b, block.AddBalance)
			err = t.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
				Key: "TotalBalance", Value: b.Text(10),
			}).Error
			if err != nil {
				return err
			}
		}

		// wormholes unique data write
		return WHInsert(t, block)
	})

	// If the write to the database is successful, update the query cache
	if err == nil {
		cache.TotalBlock++
		cache.TotalTransaction += uint64(block.TotalTransaction)
		cache.TotalUncle += uint64(block.UnclesCount)
		cache.TotalUNFT += uint64(len(block.CreateNFTs))
		cache.TotalSNFT += uint64(len(block.RewardSNFTs))
		cache.TotalSNFT -= uint64(len(block.RecycleSNFTs))
		cache.TotalExchanger += uint64(len(block.Exchangers))
		cache.TotalExchanger -= uint64(len(block.CloseExchangers))
		cache.TotalNFTTx -= uint64(len(block.NFTTxs))
		cache.TotalSNFTCollection += uint64(len(block.Epochs) * 16)
		if block.AddBalance.Uint64() != 0 {
			if block.Number == 0 {
				cache.GenesisBalance = types.BigInt(block.AddBalance.Text(10))
			}
			b := new(big.Int)
			b.SetString(string(cache.TotalBalance), 10)
			b.Add(b, block.AddBalance)
			cache.TotalBalance = types.BigInt(b.Text(10))
		}
	}
	return err
}

// SaveTxLog writes the block transaction log and analyzes and stores the ERC token transaction
func SaveTxLog(tx *gorm.DB, cacheLog []*model.Log) error {
	for _, cacheLog := range cacheLog {
		// write transaction log
		if err := tx.Create(cacheLog).Error; err != nil {
			return err
		}
		// Parse and write ERC contract transfer event
		account := model.Account{Address: cacheLog.Address}
		err := DB.Find(&account).Error
		if err != nil {
			return err
		}
		switch account.ERC {
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
	// exchange creation
	if wh.Exchangers != nil {
		err = tx.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(wh.Exchangers).Error
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
	// SNFT reward
	if wh.RewardSNFTs != nil {
		err = tx.Create(wh.RewardSNFTs).Error
		if err != nil {
			return
		}
	}
	// UNFT creation
	err = SaveUNFT(tx, wh.Number, wh.CreateNFTs)
	if err != nil {
		return
	}
	// NFT transactions, including user and official types of NFTs
	for _, nftTx := range wh.NFTTxs {
		err = SaveNFTTx(tx, nftTx)
		if err != nil {
			return
		}
	}
	// Recycle SNFT
	for _, snft := range wh.RecycleSNFTs {
		err = tx.Delete(model.SNFT{}, "address=?", snft).Error
		if err != nil {
			return
		}
	}
	for _, exchanger := range wh.CloseExchangers {
		err = tx.Delete(model.Exchanger{}, "address=?", exchanger).Error
		if err != nil {
			return
		}
	}
	// Exchange pledge
	for _, pledge := range wh.ExchangerPledges {
		err = ExchangerPledgeAdd(tx, pledge.Address, pledge.Amount)
		if err != nil {
			return
		}
	}
	// Consensus pledge
	for _, pledge := range wh.ConsensusPledges {
		err = ConsensusPledgeAdd(tx, pledge.Address, pledge.Amount)
		if err != nil {
			return
		}
	}
	if wh.Rewards != nil {
		err = tx.Create(wh.Rewards).Error
	}
	return
}

// SaveUNFT saves the NFT created by the user
func SaveUNFT(tx *gorm.DB, number types.Uint64, nfts []*model.UNFT) error {
	for i, nft := range nfts {
		*nft.Address = getNFTAddr(big.NewInt(int64(i)))
		err := tx.Create(nft).Error
		if err != nil {
			return err
		}
		// Update the total number of NFTs on the specified exchange
		if nft.ExchangerAddr != "" {
			var exchanger model.Exchanger
			err := tx.Find(&exchanger, "address=?", nft.ExchangerAddr).Error
			if err != nil {
				return err
			}
			// todo may not exist on exchange
			if exchanger.Address == nft.ExchangerAddr {
				exchanger.NFTCount++
				err = tx.Select("nft_count").Updates(&exchanger).Error
				if err != nil {
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
		// handle UNFT
		var nft model.UNFT
		err := tx.Where("address=?", nt.NFTAddr).First(&nft).Error
		if err != nil {
			return err
		}
		// populate seller field (if none)
		if nt.From == "" {
			nt.From = nft.Owner
		}
		err = tx.Model(&model.UNFT{}).Where("address=?", nft.Address).Updates(map[string]interface{}{
			"last_price": nt.Price,
			"owner":      nt.To,
		}).Error
	} else {
		// todo processing system SNFT
		var nft model.SNFT
		err := tx.Where("address=?", nt.NFTAddr).First(&nft).Error
		if err != nil {
			return err
		}
		// populate seller field (if none)
		if nt.From == "" && nft.Owner != nil {
			nt.From = *nft.Owner
		}
		err = tx.Model(&model.SNFT{}).Where("address=?", nft.Address).Updates(map[string]interface{}{
			"last_price": nt.Price,
			"owner":      nt.To,
		}).Error
	}
	// Calculate the total number of transactions and the total transaction amount to fill the NFT transaction fee and save the exchange
	if nt.ExchangerAddr != "" && nt.Price != nil && *nt.Price != "0" {
		var exchanger model.Exchanger
		err := tx.Find(&exchanger, "address=?", nt.ExchangerAddr).Error
		if err != nil {
			return err
		}
		// todo may not exist on exchange
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
			fullNFTId := collectionId + hexJ
			if epoch.Dir != "" {
				metaUrl = epoch.Dir + hexI + hexJ
			}
			// write complete SNFT information
			err = tx.Create(&model.FullNFT{ID: fullNFTId, MetaUrl: metaUrl}).Error
			if err != nil {
				return
			}
			if metaUrl != "" {
				go saveSNFTMeta(fullNFTId, metaUrl)
			}
		}
	}
	return
}

// ExchangerPledgeAdd increases the pledge amount (reduces if the amount is negative)
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

// ConsensusPledgeAdd increases the pledge amount (decreases if the amount is negative)
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

func saveSNFTCollection(id, metaUrl string) {
	nftMeta, err := GetNFTMeta(metaUrl)
	if err != nil {
		return
	}
	err = DB.Model(&model.Collection{}).Where("id=?", id).Updates(map[string]interface{}{
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
	err = DB.Model(&model.FullNFT{}).Where("id=?", id).Updates(map[string]interface{}{
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
	err = DB.Model(&model.UNFT{}).Where("address=?", nftAddr).Updates(map[string]interface{}{
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
		cache.TotalUNFTCollection += uint64(result.RowsAffected)
	}
}
