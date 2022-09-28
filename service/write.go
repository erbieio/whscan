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
	TotalTransferTx     uint64       `json:"totalTransferTx"`     //Total number of  transfer transactions
	TotalWormholesTx    uint64       `json:"totalWormholesTx"`    //Total number of  wormholes transactions
	TotalUncle          uint64       `json:"totalUncle"`          //Number of total uncle blocks
	TotalAccount        uint64       `json:"totalAccount"`        //Total account number
	GenesisBalance      types.BigInt `json:"genesisBalance"`      //Total amount of coins created
	TotalBalance        types.BigInt `json:"totalBalance"`        //The total amount of coins in the chain
	TotalExchanger      uint64       `json:"totalExchanger"`      //Total number of exchanges
	TotalNFTCollection  uint64       `json:"totalNFTCollection"`  //Total number of NFT collections
	TotalSNFTCollection uint64       `json:"totalSNFTCollection"` //Total number of SNFT collections
	TotalNFT            uint64       `json:"totalNFT"`            //Total number of NFTs
	TotalSNFT           uint64       `json:"totalSNFT"`           //Total number of SNFTs
	TotalNFTTx          uint64       `json:"totalNFTTx"`          //Total number of  NFT transactions
	TotalSNFTTx         uint64       `json:"totalSNFTTx"`         //Total number of  SNFT transactions
	TotalAmount         types.BigInt `json:"totalAmount"`         //total transaction volume
	TotalNFTAmount      types.BigInt `json:"totalNFTAmount"`      //Total transaction volume of NFTs
	TotalSNFTAmount     types.BigInt `json:"totalSNFTAmount"`     //Total transaction volume of SNFTs
	TotalNFTCreator     uint64       `json:"totalNFTCreator"`     //Total creator of NFTs
	TotalSNFTCreator    uint64       `json:"totalSNFTCreator"`    //Total creator of SNFTs
	Total24HTx          uint64       `json:"total24HTx"`          //Total number of transactions within 24 hours
	TotalExchangerTx    uint64       `json:"totalExchangerTx"`    //Total number of exchanger  transactions
	Total24HExchangerTx uint64       `json:"total24HExchangerTx"` //Total number of exchanger  transactions within 24 hours
	Total24HNFT         uint64       `json:"total24HNFT"`         //Total number of NFT within 24 hours
	RewardCoinCount     uint64       `json:"rewardCoinCount"`     //Total number of times to get coin rewards, 0.1ERB once
	RewardSNFTCount     uint64       `json:"rewardSNFTCount"`     //Total number of times to get SNFT rewards
	TotalRecycle        uint64       `json:"totalRecycle"`        //Total number of recycle SNFT
	TotalValidator      uint64       `json:"totalValidator"`      // Total number of validator
	TotalPledge         types.BigInt `json:"totalPledge"`         //Total amount of validator pledge
}

var cache = Cache{
	TotalAmount:     "0",
	TotalNFTAmount:  "0",
	TotalSNFTAmount: "0",
	TotalPledge:     "0",
}

var fnfts = make(map[string]int64)

// InitCache initializes the query cache from the database
func initCache() (err error) {
	if err = DB.Model(&model.Block{}).Select("COUNT(*)").Scan(&cache.TotalBlock).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Transaction{}).Select("COUNT(*)").Scan(&cache.TotalTransaction).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Transaction{}).Select("COUNT(*)").Where("input='0x'").Scan(&cache.TotalTransferTx).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Transaction{}).Select("COUNT(*)").Where("LEFT(input,22)='0x776f726d686f6c65733a'").Scan(&cache.TotalWormholesTx).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Uncle{}).Select("COUNT(*)").Scan(&cache.TotalUncle).Error; err != nil {
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
	if err = DB.Model(&model.Collection{}).Where("length(id)!=40").Select("COUNT(*)").Scan(&cache.TotalNFTCollection).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Collection{}).Where("length(id)=40").Select("COUNT(*)").Scan(&cache.TotalSNFTCollection).Error; err != nil {
		return
	}
	if err = DB.Model(&model.NFT{}).Select("COUNT(*)").Scan(&cache.TotalNFT).Error; err != nil {
		return
	}
	if err = DB.Model(&model.SNFT{}).Select("COUNT(*)").Scan(&cache.TotalSNFT).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Reward{}).Select("COUNT(snft)").Scan(&cache.RewardSNFTCount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Reward{}).Select("COUNT(amount)").Scan(&cache.RewardCoinCount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x0'").Select("COUNT(*)").Scan(&cache.TotalNFTTx).Error; err != nil {
		return
	}
	if err = DB.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x8'").Select("COUNT(*)").Scan(&cache.TotalSNFTTx).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalAmount").Select("value").Scan(&cache.TotalAmount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalNFTAmount").Select("value").Scan(&cache.TotalNFTAmount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalSNFTAmount").Select("value").Scan(&cache.TotalSNFTAmount).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalRecycle").Select("value").Scan(&cache.TotalRecycle).Error; err != nil {
		return
	}
	if err = DB.Model(&model.Cache{}).Where("`key`=?", "TotalPledge").Select("value").Scan(&cache.TotalPledge).Error; err != nil {
		return
	}
	freshCache()
	return err
}

func TotalBlock() uint64 {
	return cache.TotalBlock
}

var lastTime = int64(0)

func freshCache() {
	if now := time.Now().Unix(); now-lastTime > 60 {
		var number uint64
		if err := DB.Model(&model.Pledge{}).Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.TotalValidator = number
		}
		if err := DB.Model(&model.Account{}).Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.TotalAccount = number
		}
		if err := DB.Model(&model.NFT{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			cache.TotalNFTCreator = number
		}
		if err := DB.Model(&model.Epoch{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			cache.TotalSNFTCreator = number
		}
		if err := DB.Model(&model.Block{}).Where("timestamp>?", now-86400).Select("IFNULL(SUM(total_transaction),0)").Scan(&number).Error; err == nil {
			cache.Total24HTx = number
		}
		if err := DB.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL").Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.TotalExchangerTx = number
		}
		if err := DB.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL AND timestamp>?", now-86400).Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.Total24HExchangerTx = number
		}
		if err := DB.Model(&model.NFT{}).Where("timestamp>?", now-86400).Select("COUNT(*)").Scan(&number).Error; err == nil {
			cache.Total24HNFT = number
		}
		for fnft := range fnfts {
			err := DB.Exec("CAll fresh_c_snft(?)", fnft).Error
			if err != nil {
				log.Println("fresh com-snft error:", err)
			}
		}
		fnfts = make(map[string]int64)
		lastTime = now
	}
}

func TotalBalance() types.BigInt {
	if cache.TotalBalance == "" {
		cache.TotalBalance = "0"
	}
	return cache.TotalBalance
}

// getNFTAddr Get the NFT address
func getNFTAddr(next *big.Int) string {
	return string(utils.BigToAddress(next.Add(next, big.NewInt(int64(cache.TotalNFT+1)))))
}

func BlockInsert(block *model.Parsed) error {
	totalAmount, totalPledge, b := new(big.Int), new(big.Int), new(big.Int)
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
		for _, log := range block.CacheLogs {
			if err := t.Create(log).Error; err != nil {
				return err
			}
		}
		for _, log := range block.CacheTransferLogs {
			if err := t.Create(log).Error; err != nil {
				return err
			}
		}

		// write block
		if err := t.Create(block.Block).Error; err != nil {
			return err
		}

		if block.AddBalance.Uint64() != 0 {
			if block.Number == 0 {
				err := t.Create(model.Cache{
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
			err := t.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
				Key: "TotalBalance", Value: b.Text(10),
			}).Error
			if err != nil {
				return err
			}
		}
		err := t.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
			Key: "TotalRecycle", Value: fmt.Sprintf("%d", cache.TotalRecycle+uint64(len(block.RecycleSNFTs))),
		}).Error
		if err != nil {
			return err
		}
		if len(block.CacheTxs) > 0 {
			totalAmount.SetString(string(cache.TotalAmount), 10)
			for _, tx := range block.CacheTxs {
				b.SetString(string(tx.Value), 10)
				totalAmount = totalAmount.Add(totalAmount, b)
			}
			err = t.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(&model.Cache{
				Key: "TotalAmount", Value: totalAmount.Text(10)}).Error
			if err != nil {
				return err
			}
		}
		if len(block.ConsensusPledges) > 0 {
			totalPledge.SetString(string(cache.TotalPledge), 10)
			for _, pledge := range block.ConsensusPledges {
				b.SetString(pledge.Amount, 10)
				totalPledge = totalPledge.Add(totalPledge, b)
			}
			err = t.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(&model.Cache{
				Key: "TotalPledge", Value: totalPledge.Text(10)}).Error
			if err != nil {
				return err
			}
		}

		// wormholes unique data write
		return WHInsert(t, block)
	})

	// If write to the database is successful, update the query cache
	if err == nil {
		cache.TotalBlock++
		cache.TotalTransaction += uint64(block.TotalTransaction)
		cache.TotalUncle += uint64(block.UnclesCount)
		cache.TotalNFT += uint64(len(block.CreateNFTs))
		cache.TotalSNFT += uint64(len(block.RewardSNFTs))
		cache.TotalSNFT -= uint64(len(block.RecycleSNFTs))
		cache.RewardSNFTCount += uint64(len(block.RewardSNFTs))
		cache.RewardCoinCount += uint64(len(block.Rewards) - len(block.RewardSNFTs))
		cache.TotalExchanger += uint64(len(block.Exchangers))
		cache.TotalExchanger -= uint64(len(block.CloseExchangers))
		cache.TotalRecycle += uint64(len(block.RecycleSNFTs))

		for _, tx := range block.NFTTxs {
			if (*tx.NFTAddr)[:3] == "0x0" {
				cache.TotalNFTTx += 1
				if tx.Price != nil {
					b, price := new(big.Int), new(big.Int)
					b.SetString(string(cache.TotalNFTAmount), 10)
					price.SetString(*tx.Price, 10)
					b = b.Add(b, price)
					cache.TotalNFTAmount = types.BigInt(b.Text(10))
				}
			} else {
				cache.TotalSNFTTx += 1
				if tx.Price != nil {
					b, price := new(big.Int), new(big.Int)
					b.SetString(string(cache.TotalSNFTAmount), 10)
					price.SetString(*tx.Price, 10)
					b = b.Add(b, price)
					cache.TotalSNFTAmount = types.BigInt(b.Text(10))
				}
				fnfts[(*tx.NFTAddr)[:41]] = 0
			}
		}
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
		for _, tx := range block.CacheTxs {
			if tx.Input == "0x" {
				cache.TotalTransferTx++
			} else if len(tx.Input) > 22 && tx.Input[:22] == "0x776f726d686f6c65733a" {
				cache.TotalWormholesTx++
			}
		}
		if len(block.CacheTxs) > 0 {
			cache.TotalAmount = types.BigInt(totalAmount.Text(10))
		}
		if len(block.ConsensusPledges) > 0 {
			cache.TotalPledge = types.BigInt(totalPledge.Text(10))
		}
		for _, snft := range block.RewardSNFTs {
			fnfts[snft.Address[:41]] = 0
		}
		for _, snft := range block.PledgeSNFT {
			fnfts[snft[:41]] = 0
		}
		for _, snft := range block.UnPledgeSNFT {
			fnfts[snft[:41]] = 0
		}
		for _, snft := range block.RecycleSNFTs {
			fnfts[snft[:41]] = 0
		}
		freshCache()
	}
	return err
}

func WHInsert(tx *gorm.DB, wh *model.Parsed) (err error) {
	// exchange creation
	if wh.Exchangers != nil {
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"name", "url", "fee_ratio", "timestamp", "block_number", "tx_hash", "close_at"}),
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
	for _, snft := range wh.RecycleSNFTs {
		err = tx.Delete(model.SNFT{}, "address=?", snft).Error
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
	// NFT creation
	err = SaveNFT(tx, wh.Number, wh.CreateNFTs)
	if err != nil {
		return
	}
	// NFT transactions, including user and official types of NFTs
	for _, nftTx := range wh.NFTTxs {
		err = SaveNFTTx(tx, nftTx)
		if err != nil {
			return
		}
		if (*nftTx.NFTAddr)[:3] == "0x0" {
			if nftTx.Price != nil {
				b, price := new(big.Int), new(big.Int)
				b.SetString(string(cache.TotalNFTAmount), 10)
				price.SetString(*nftTx.Price, 10)
				b = b.Add(b, price)
				err = tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
					Key: "TotalNFTAmount", Value: b.Text(10),
				}).Error
				if err != nil {
					return err
				}
			}
		} else {
			if nftTx.Price != nil {
				b, price := new(big.Int), new(big.Int)
				b.SetString(string(cache.TotalSNFTAmount), 10)
				price.SetString(*nftTx.Price, 10)
				b = b.Add(b, price)
				err = tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
					Key: "TotalSNFTAmount", Value: b.Text(10),
				}).Error
				if err != nil {
					return err
				}
			}
		}
	}
	// exchanger pledge
	for _, pledge := range wh.ExchangerPledges {
		err = ExchangerPledgeAdd(tx, pledge.Address, pledge.Amount)
		if err != nil {
			return
		}
	}
	// close exchanger
	for _, exchanger := range wh.CloseExchangers {
		err = tx.Model(model.Exchanger{}).Where("address=?", exchanger).Updates(map[string]interface{}{"close_at": wh.Timestamp, "amount": "0"}).Error
		if err != nil {
			return
		}
	}
	// consensus pledge
	for _, pledge := range wh.ConsensusPledges {
		var _pledge model.Pledge
		if err = tx.Find(&_pledge, "address=?", pledge.Address).Error; err != nil {
			return
		}
		if len(_pledge.Address) == 0 {
			_pledge.Address = pledge.Address
			_pledge.Amount = "0"
			_pledge.Reward = "0"
		}
		_pledge.Amount = BigIntAdd(_pledge.Amount, pledge.Amount)
		_pledge.Count++
		_pledge.Timestamp = uint64(wh.Timestamp)
		_pledge.LastNumber = uint64(wh.Number)
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"amount", "count", "timestamp", "last_number"}),
		}).Create(&_pledge).Error
		if err != nil {
			return
		}
	}
	for _, snft := range wh.PledgeSNFT {
		err = tx.Exec("UPDATE snfts SET pledge_number=? WHERE address=?", wh.Number, snft).Error
		if err != nil {
			return
		}
	}
	for _, snft := range wh.UnPledgeSNFT {
		err = tx.Exec("UPDATE snfts SET pledge_number=NULL WHERE address=?", snft).Error
		if err != nil {
			return
		}
	}
	for _, reward := range wh.Rewards {
		if err = tx.Create(reward).Error; err != nil {
			return
		}
		if reward.Identity == 2 {
			var pledge model.Pledge
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

// SaveNFT saves the NFT created by the user
func SaveNFT(tx *gorm.DB, number types.Uint64, nfts []*model.NFT) error {
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
		err = tx.Model(&model.NFT{}).Where("address=?", nft.Address).Updates(map[string]interface{}{
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
		if nt.From == "" {
			nt.From = nft.Owner
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
		if exchanger.Address == nt.ExchangerAddr {
			price, _ := new(big.Int).SetString(*nt.Price, 10)
			fee := big.NewInt(int64(exchanger.FeeRatio))
			fee = fee.Mul(fee, price)
			feeStr := fee.Div(fee, big.NewInt(10000)).Text(10)
			nt.Fee = &feeStr
			balanceCount := new(big.Int)
			balanceCount.SetString(exchanger.BalanceCount, 10)
			balanceCount = balanceCount.Add(balanceCount, price)
			exchanger.BalanceCount = balanceCount.Text(10)
			err = tx.Select("balance_count").Updates(&exchanger).Error
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

// ExchangerPledgeAdd increases the pledge amount (reduces if the amount is negative)
func ExchangerPledgeAdd(tx *gorm.DB, addr, amount string) error {
	pledge := model.Exchanger{}
	err := tx.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		pledge.Amount = "0"
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
	err = DB.Model(&model.FNFT{}).Where("id=?", id).Updates(map[string]interface{}{
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
	err = DB.Model(&model.NFT{}).Where("address=?", nftAddr).Updates(map[string]interface{}{
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
		cache.TotalNFTCollection += uint64(result.RowsAffected)
	}
}
