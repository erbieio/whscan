package service

import (
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// Stats caches some database queries to speed up queries
type Stats struct {
	ChainId              int64  `json:"chainId"`              //chain id
	GenesisBalance       string `json:"genesisBalance"`       //Total amount of coins created
	AvgBlockTime         int64  `json:"avgBlockTime"`         //average block time, ms
	TotalBlock           int64  `json:"totalBlock"`           //Total number of blocks
	TotalBlackHole       int64  `json:"totalBlackHole"`       //Total number of BlackHole blocks
	TotalTransaction     int64  `json:"totalTransaction"`     //Total number of transactions
	TotalInternalTx      int64  `json:"totalInternalTx"`      //Total number of internal transactions
	TotalTransferTx      int64  `json:"totalTransferTx"`      //Total number of  transfer transactions
	TotalWormholesTx     int64  `json:"totalWormholesTx"`     //Total number of  wormholes transactions
	TotalUncle           int64  `json:"totalUncle"`           //Number of total uncle blocks
	TotalAccount         int64  `json:"totalAccount"`         //Total account number
	TotalBalance         string `json:"totalBalance"`         //The total amount of coins in the chain
	TotalExchanger       int64  `json:"totalExchanger"`       //Total number of exchanges
	TotalNFTCollection   int64  `json:"totalNFTCollection"`   //Total number of NFT collections
	TotalSNFTCollection  int64  `json:"totalSNFTCollection"`  //Total number of SNFT collections
	TotalNFT             int64  `json:"totalNFT"`             //Total number of NFTs
	TotalSNFT            int64  `json:"totalSNFT"`            //Total number of SNFTs
	TotalNFTTx           int64  `json:"totalNFTTx"`           //Total number of  NFT transactions
	TotalSNFTTx          int64  `json:"totalSNFTTx"`          //Total number of  SNFT transactions
	TotalAmount          string `json:"totalAmount"`          //total transaction volume
	TotalNFTAmount       string `json:"totalNFTAmount"`       //Total transaction volume of NFTs
	TotalSNFTAmount      string `json:"totalSNFTAmount"`      //Total transaction volume of SNFTs
	TotalValidator       int64  `json:"totalValidator"`       //Total number of validator
	TotalNFTCreator      int64  `json:"totalNFTCreator"`      //Total creator of NFTs
	TotalSNFTCreator     int64  `json:"totalSNFTCreator"`     //Total creator of SNFTs
	TotalExchangerTx     int64  `json:"totalExchangerTx"`     //Total number of exchanger  transactions
	RewardCoinCount      int64  `json:"rewardCoinCount"`      //Total number of times to get coin rewards, 0.1ERB once
	RewardSNFTCount      int64  `json:"rewardSNFTCount"`      //Total number of times to get SNFT rewards
	TotalRecycle         uint64 `json:"totalRecycle"`         //Total number of recycle SNFT
	TotalValidatorPledge string `json:"totalValidatorPledge"` //Total amount of validator pledge
	TotalExchangerPledge string `json:"totalExchangerPledge"` //Total amount of exchanger pledge
	TotalSNFTPledge      string `json:"totalSNFTPledge"`      //Total amount of snft pledge
	Total24HExchangerTx  int64  `json:"total24HExchangerTx"`  //Total number of exchanger  transactions within 24 hours
	Total24HNFT          int64  `json:"total24HNFT"`          //Total number of NFT within 24 hours
	Total24HTx           int64  `json:"total24HTx"`           //Total number of transactions within 24 hours

	genesis    model.Header
	firstBlock model.Header
	balances   map[types.Address]*big.Int
}

var stats = Stats{
	TotalAmount:     "0",
	TotalNFTAmount:  "0",
	TotalSNFTAmount: "0",
	balances:        make(map[types.Address]*big.Int),
}

// initStats initializes the query stats from the database
func initStats(db *gorm.DB) (err error) {
	var accounts []*model.Account
	if err = db.Select("address", "balance").Find(&accounts).Error; err != nil {
		return
	}
	for _, account := range accounts {
		stats.balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
	}
	return loadStats(db)
}

func loadStats(db *gorm.DB) (err error) {
	if err = db.Model(&model.Cache{}).Where("`key`='ChainId'").Pluck("value", &stats.ChainId).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`='GenesisBalance'").Pluck("value", &stats.GenesisBalance).Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Count(&stats.TotalBlock).Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Where("`number`>0 AND `miner`='0x0000000000000000000000000000000000000000'").Count(&stats.TotalBlackHole).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Count(&stats.TotalTransaction).Error; err != nil {
		return
	}
	if err = db.Model(&model.InternalTx{}).Count(&stats.TotalInternalTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Where("input='0x'").Count(&stats.TotalTransferTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Where("LEFT(input,22)='0x776f726d686f6c65733a'").Count(&stats.TotalWormholesTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Uncle{}).Count(&stats.TotalUncle).Error; err != nil {
		return
	}
	if err = db.Model(&model.Collection{}).Where("length(id)=40").Count(&stats.TotalSNFTCollection).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFT{}).Count(&stats.TotalNFT).Error; err != nil {
		return
	}
	if err = db.Model(&model.SNFT{}).Count(&stats.TotalSNFT).Error; err != nil {
		return
	}
	if err = db.Model(&model.Reward{}).Select("COUNT(snft)").Scan(&stats.RewardSNFTCount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Reward{}).Select("COUNT(amount)").Scan(&stats.RewardCoinCount).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x0'").Count(&stats.TotalNFTTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x8'").Count(&stats.TotalSNFTTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`='TotalAmount'").Pluck("value", &stats.TotalAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`='TotalNFTAmount'").Pluck("value", &stats.TotalNFTAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`='TotalSNFTAmount'").Pluck("value", &stats.TotalSNFTAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalRecycle").Select("value").Scan(&stats.TotalRecycle).Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Find(&stats.genesis, "number=0").Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Find(&stats.firstBlock, "number=1").Error; err != nil {
		return
	}
	stats.TotalAccount = int64(len(stats.balances))
	totalBalance := new(big.Int)
	for _, balance := range stats.balances {
		totalBalance = totalBalance.Add(totalBalance, balance)
	}
	stats.TotalBalance = totalBalance.Text(10)

	value, totalSNFTPledge, snftAmounts := new(big.Int), new(big.Int), make([]string, 0)
	if err = db.Model(&model.Account{}).Where("`snft_amount`!='0'").Pluck("snft_amount", &snftAmounts).Error; err != nil {
		return
	}
	for _, snftAmount := range snftAmounts {
		value.SetString(snftAmount, 0)
		totalSNFTPledge = totalSNFTPledge.Add(totalSNFTPledge, value)
	}
	stats.TotalSNFTPledge = totalSNFTPledge.Text(10)

	totalExchangerPledge, exchangerAmounts := new(big.Int), make([]string, 0)
	if err = db.Model(&model.Exchanger{}).Where("`amount`!='0'").Pluck("amount", &exchangerAmounts).Error; err != nil {
		return
	}
	for _, exchangerAmount := range exchangerAmounts {
		value.SetString(exchangerAmount, 0)
		totalExchangerPledge = totalExchangerPledge.Add(totalExchangerPledge, value)
	}
	stats.TotalExchangerPledge = totalExchangerPledge.Text(10)

	totalValidatorPledge, validatorAmounts := new(big.Int), make([]string, 0)
	if err = db.Model(&model.Validator{}).Where("`amount`!='0'").Pluck("amount", &validatorAmounts).Error; err != nil {
		return
	}
	for _, validatorAmount := range validatorAmounts {
		value.SetString(validatorAmount, 0)
		totalValidatorPledge = totalValidatorPledge.Add(totalValidatorPledge, value)
	}
	stats.TotalValidatorPledge = totalValidatorPledge.Text(10)
	return
}

func updateStats(db *gorm.DB, parsed *model.Parsed) (err error) {
	totalBalance, _ := new(big.Int).SetString(stats.TotalBalance, 0)
	totalAmount, _ := new(big.Int).SetString(stats.TotalAmount, 0)
	totalValidatorPledge, _ := new(big.Int).SetString(stats.TotalValidatorPledge, 0)
	totalExchangerPledge, _ := new(big.Int).SetString(stats.TotalExchangerPledge, 0)
	totalSNFTPledge, _ := new(big.Int).SetString(stats.TotalSNFTPledge, 0)
	totalNFTAmount, _ := new(big.Int).SetString(stats.TotalNFTAmount, 0)
	totalSNFTAmount, _ := new(big.Int).SetString(stats.TotalSNFTAmount, 0)
	rewardSNFT, value, totalNFTTx, totalSNFTTx := 0, new(big.Int), stats.TotalNFTTx, stats.TotalSNFTTx
	fnfts := make(map[string]int64)
	for _, account := range parsed.CacheAccounts {
		value.SetString(string(account.Balance), 0)
		totalBalance = totalBalance.Add(totalBalance, value)
		if stats.balances[account.Address] != nil {
			totalBalance = totalBalance.Sub(totalBalance, stats.balances[account.Address])
		}
	}
	for _, tx := range parsed.CacheTxs {
		value.SetString(string(tx.Value), 0)
		totalAmount = totalAmount.Add(totalAmount, value)
	}
	for _, pledge := range parsed.ChangeValidators {
		value.SetString(pledge.Amount, 0)
		totalValidatorPledge = totalValidatorPledge.Add(totalValidatorPledge, value)
	}
	for _, pledge := range parsed.ChangeExchangers {
		value.SetString(pledge.Amount, 0)
		totalExchangerPledge = totalExchangerPledge.Add(totalExchangerPledge, value)
	}
	for _, tx := range parsed.NFTTxs {
		if (*tx.NFTAddr)[:3] == "0x0" {
			totalNFTTx++
			if tx.Price != "0" {
				value.SetString(tx.Price, 0)
				totalNFTAmount = totalNFTAmount.Add(totalNFTAmount, value)
			}
		} else {
			totalSNFTTx++
			if tx.Price != "0" {
				value.SetString(tx.Price, 0)
				totalSNFTAmount = totalSNFTAmount.Add(totalSNFTAmount, value)
			}
			fnfts[(*tx.NFTAddr + "00")[:41]] = 0
		}
	}
	for _, reward := range parsed.Rewards {
		if reward.SNFT != nil {
			fnfts[(*reward.SNFT)[:41]] = 0
			rewardSNFT++
		}
	}
	for _, snft := range parsed.PledgeSNFT {
		fnfts[(snft + "00")[42:83]] = 0
	}
	for _, snft := range parsed.UnPledgeSNFT {
		fnfts[(snft + "00")[42:83]] = 0
	}
	for _, snft := range parsed.RecycleSNFTs {
		fnfts[(snft + "00")[:41]] = 0
	}
	for fnft := range fnfts {
		err = db.Exec("CAll fresh_c_snft(?)", fnft).Error
		if err != nil {
			return
		}
	}
	if parsed.Number == 0 {
		if err = db.Create(&model.Cache{Key: "ChainId", Value: fmt.Sprintf("%v", stats.ChainId)}).Error; err != nil {
			return
		}
		if err = db.Create(&model.Cache{Key: "GenesisBalance", Value: totalBalance.Text(10)}).Error; err != nil {
			return
		}
		stats.GenesisBalance = totalBalance.Text(10)
	}
	if count := len(parsed.RecycleSNFTs); count > 0 {
		err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(model.Cache{
			Key: "TotalRecycle", Value: fmt.Sprintf("%d", stats.TotalRecycle+uint64(count)),
		}).Error
		if err != nil {
			return
		}
	}
	if len(parsed.CacheTxs) > 0 {
		err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(&model.Cache{
			Key: "TotalAmount", Value: totalAmount.Text(10)}).Error
		if err != nil {
			return
		}
	}
	if len(parsed.NFTTxs) > 0 {
		err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(&model.Cache{
			Key: "TotalNFTAmount", Value: totalNFTAmount.Text(10)}).Error
		if err != nil {
			return
		}
		err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"value"})}).Create(&model.Cache{
			Key: "TotalSNFTAmount", Value: totalSNFTAmount.Text(10)}).Error
		if err != nil {
			return
		}
	}
	for _, account := range parsed.CacheAccounts {
		stats.balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
	}
	if parsed.Number == 1 {
		stats.firstBlock = parsed.Block.Header
	}
	if parsed.Number > 1 {
		stats.AvgBlockTime = int64((parsed.Timestamp-stats.firstBlock.Timestamp)*1000/parsed.Number - 1)
	}
	stats.TotalBlock++
	stats.TotalTransaction += int64(len(parsed.CacheTxs))
	stats.TotalInternalTx += int64(len(parsed.CacheInternalTxs))
	stats.TotalUncle += int64(parsed.UnclesCount)
	stats.TotalNFT += int64(len(parsed.NFTs))
	stats.TotalSNFT += int64(rewardSNFT - len(parsed.RecycleSNFTs))
	stats.RewardSNFTCount += int64(rewardSNFT)
	stats.RewardCoinCount += int64(len(parsed.Rewards) - rewardSNFT)
	stats.TotalAccount = int64(len(stats.balances))
	stats.TotalRecycle += uint64(len(parsed.RecycleSNFTs))
	stats.TotalBalance = totalBalance.Text(10)
	stats.TotalAmount = totalAmount.Text(10)
	stats.TotalValidatorPledge = totalValidatorPledge.Text(10)
	stats.TotalExchangerPledge = totalExchangerPledge.Text(10)
	stats.TotalSNFTPledge = totalSNFTPledge.Text(10)
	stats.TotalNFTTx = totalNFTTx
	stats.TotalSNFTTx = totalSNFTTx
	stats.TotalNFTAmount = totalNFTAmount.Text(10)
	stats.TotalSNFTAmount = totalSNFTAmount.Text(10)
	stats.TotalSNFTCollection = (stats.RewardSNFTCount/4096 + 1) * 16
	if parsed.Number > 0 && parsed.Miner == "0x0000000000000000000000000000000000000000" {
		stats.TotalBlackHole++
	}
	for _, tx := range parsed.CacheTxs {
		if tx.Input == "0x" {
			stats.TotalTransferTx++
		} else if len(tx.Input) > 22 && tx.Input[:22] == "0x776f726d686f6c65733a" {
			stats.TotalWormholesTx++
		}
	}
	return
}

func fixStats(db *gorm.DB, parsed *model.Parsed) (err error) {
	for address := range stats.balances {
		if account := parsed.CacheAccounts[address]; account != nil {
			if err = db.Select("balance", "nonce").Updates(account).Error; err != nil {
				return
			}
			stats.balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
		} else {
			if err = db.Delete(&model.Account{}, "address=?", address).Error; err != nil {
				return
			}
			delete(stats.balances, address)
		}
	}
	return loadStats(db)
}

var lastTime time.Time

func freshStats(db *gorm.DB) {
	if now := time.Now(); now.Minute() != lastTime.Minute() {
		var number int64
		if err := db.Model(&model.Exchanger{}).Count(&number).Error; err == nil {
			stats.TotalExchanger = number
		}
		if err := db.Model(&model.Collection{}).Where("length(id)!=40").Count(&number).Error; err == nil {
			stats.TotalNFTCollection = number
		}
		if err := db.Model(&model.Validator{}).Count(&number).Error; err == nil {
			stats.TotalValidator = number
		}
		if err := db.Model(&model.NFT{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			stats.TotalNFTCreator = number
		}
		if err := db.Model(&model.Epoch{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			stats.TotalSNFTCreator = number
		}
		if err := db.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL").Count(&number).Error; err == nil {
			stats.TotalExchangerTx = number
		}
		if now.Hour() != lastTime.Hour() {
			start, stop := utils.LastTimeRange(1)
			if err := db.Model(&model.Block{}).Where("timestamp>=? AND timestamp<?", start, stop).Select("IFNULL(SUM(total_transaction),0)").Scan(&number).Error; err == nil {
				stats.Total24HTx = number
			}
			if err := db.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL AND timestamp>=? AND timestamp<?", start, stop).Select("COUNT(*)").Scan(&number).Error; err == nil {
				stats.Total24HExchangerTx = number
			}
			if err := db.Model(&model.NFT{}).Where("timestamp>=? AND timestamp<?", start, stop).Select("COUNT(*)").Scan(&number).Error; err == nil {
				stats.Total24HNFT = number
			}
		}
		lastTime = now
	}
}

func CheckStats(chainId types.Uint64, genesis *model.Header) bool {
	if stats.TotalBlock == 0 {
		stats.ChainId = int64(chainId)
		stats.genesis = *genesis
	} else if stats.ChainId != int64(chainId) || stats.genesis != *genesis {
		return false
	}
	return true
}

func TotalBlock() types.Uint64 {
	return types.Uint64(stats.TotalBlock)
}
