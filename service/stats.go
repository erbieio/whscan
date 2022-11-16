package service

import (
	"math/big"
	"time"

	"gorm.io/gorm"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

var stats = &model.Stats{
	TotalAmount:     "0",
	TotalNFTAmount:  "0",
	TotalSNFTAmount: "0",
	Balances:        make(map[types.Address]*big.Int),
}

// initStats initializes the query stats from the database
func initStats(db *gorm.DB) (err error) {
	var accounts []*model.Account
	if err = db.Select("address", "balance").Find(&accounts).Error; err != nil {
		return
	}
	for _, account := range accounts {
		stats.Balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
	}
	return loadStats(db)
}

func loadStats(db *gorm.DB) (err error) {
	if err = db.Model(&model.Stats{}).Scan(&stats).Error; err != nil {
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
	if err = db.Model(&model.Block{}).Find(&stats.Genesis, "number=0").Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Find(&stats.FirstBlock, "number=1").Error; err != nil {
		return
	}
	stats.TotalAccount = int64(len(stats.Balances))
	totalBalance := new(big.Int)
	for _, balance := range stats.Balances {
		totalBalance = totalBalance.Add(totalBalance, balance)
	}
	stats.TotalBalance = totalBalance.Text(10)

	value, totalSNFTPledge, snftAmounts := new(big.Int), new(big.Int), make([]string, 0)
	if err = db.Model(&model.User{}).Pluck("amount", &snftAmounts).Error; err != nil {
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
		if stats.Balances[account.Address] != nil {
			totalBalance = totalBalance.Sub(totalBalance, stats.Balances[account.Address])
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
	if stats.Ready {
		for fnft := range fnfts {
			err = db.Exec("CAll fresh_c_snft(?)", fnft).Error
			if err != nil {
				return
			}
		}
	}
	if parsed.Number == 0 {
		stats.GenesisBalance = totalBalance.Text(10)
		if err = db.Create(&stats).Error; err != nil {
			return
		}
	}
	if count := len(parsed.RecycleSNFTs); count > 0 {
		err = db.Model(&model.Stats{}).Where("`chain_id`=?", stats.ChainId).Update("total_recycle", stats.TotalRecycle+uint64(count)).Error
		if err != nil {
			return
		}
	}
	if len(parsed.CacheTxs) > 0 {
		err = db.Model(&model.Stats{}).Where("`chain_id`=?", stats.ChainId).Update("total_amount", totalAmount.Text(10)).Error
		if err != nil {
			return
		}
	}
	if len(parsed.NFTTxs) > 0 {
		err = db.Model(&model.Stats{}).Where("`chain_id`=?", stats.ChainId).Update("total_nft_amount", totalNFTAmount.Text(10)).Error
		if err != nil {
			return
		}
		err = db.Model(&model.Stats{}).Where("`chain_id`=?", stats.ChainId).Update("total_snft_amount", totalSNFTAmount.Text(10)).Error
		if err != nil {
			return
		}
	}
	for _, account := range parsed.CacheAccounts {
		stats.Balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
	}
	if parsed.Number == 1 {
		stats.FirstBlock = parsed.Block.Header
	}
	if parsed.Number > 1 {
		stats.AvgBlockTime = int64((parsed.Timestamp-stats.FirstBlock.Timestamp)*1000/parsed.Number - 1)
	}
	stats.TotalBlock++
	stats.TotalTransaction += int64(len(parsed.CacheTxs))
	stats.TotalInternalTx += int64(len(parsed.CacheInternalTxs))
	stats.TotalUncle += int64(parsed.UnclesCount)
	stats.TotalNFT += int64(len(parsed.NFTs))
	stats.TotalSNFT += int64(rewardSNFT - len(parsed.RecycleSNFTs))
	stats.RewardSNFTCount += int64(rewardSNFT)
	stats.RewardCoinCount += int64(len(parsed.Rewards) - rewardSNFT)
	stats.TotalAccount = int64(len(stats.Balances))
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
	for address := range stats.Balances {
		if account := parsed.CacheAccounts[address]; account != nil {
			if err = db.Select("balance", "nonce").Updates(account).Error; err != nil {
				return
			}
			stats.Balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
		} else {
			if err = db.Delete(&model.Account{}, "address=?", address).Error; err != nil {
				return
			}
			delete(stats.Balances, address)
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

func GetStats() *model.Stats {
	return stats
}
