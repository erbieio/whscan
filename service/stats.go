package service

import (
	"math/big"
	"strconv"

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
	if err = db.Model(&model.Transaction{}).Where("LEFT(input,14)='0x65726269653a'").Count(&stats.TotalErbieTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFT{}).Count(&stats.TotalNFT).Error; err != nil {
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
	if err = db.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL").Count(&stats.TotalStakerTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Epoch{}).Count(&stats.TotalEpoch).Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Find(&stats.Genesis, "number=0").Error; err != nil {
		return
	}
	stats.TotalAccount = int64(len(stats.Balances))
	totalBalance := new(big.Int)
	for _, balance := range stats.Balances {
		totalBalance = totalBalance.Add(totalBalance, balance)
	}
	stats.TotalBalance = totalBalance.Text(10)

	value, totalPledge, amounts := new(big.Int), new(big.Int), make([]string, 0)
	if err = db.Model(&model.Pledge{}).Pluck("amount", &amounts).Error; err != nil {
		return
	}
	for _, amount := range amounts {
		value.SetString(amount, 0)
		totalPledge = totalPledge.Add(totalPledge, value)
	}
	stats.TotalPledge = totalPledge.Text(10)
	return
}

func updateStats(db *gorm.DB, parsed *model.Parsed) (err error) {
	totalBalance, _ := new(big.Int).SetString(stats.TotalBalance, 0)
	totalAmount, _ := new(big.Int).SetString(stats.TotalAmount, 0)
	totalPledge, _ := new(big.Int).SetString(stats.TotalPledge, 0)
	totalNFTAmount, _ := new(big.Int).SetString(stats.TotalNFTAmount, 0)
	totalSNFTAmount, _ := new(big.Int).SetString(stats.TotalSNFTAmount, 0)
	rewardSNFT, recycleSNFT, value := int64(0), int64(0), new(big.Int)
	totalNFTTx, totalSNFTTx, totalStakerTx := stats.TotalNFTTx, stats.TotalSNFTTx, stats.TotalStakerTx
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
	for _, pledge := range parsed.ChangePledges {
		value.SetString(pledge.Amount, 0)
		totalPledge = totalPledge.Add(totalPledge, value)
	}
	for _, tx := range parsed.NFTTxs {
		if tx.ExchangerAddr != nil {
			totalStakerTx++
		}
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
			if tx.TxType == 6 {
				pieces, _ := strconv.ParseInt(*tx.Fee, 10, 32)
				recycleSNFT += pieces
			}
		}
	}
	for _, reward := range parsed.Rewards {
		if reward.SNFT != nil {
			rewardSNFT++
		}
	}
	if parsed.Number == 0 {
		stats.GenesisBalance = totalBalance.Text(10)
		if err = db.Create(&stats).Error; err != nil {
			return
		}
	}
	if recycleSNFT > 0 {
		err = db.Model(&model.Stats{}).Where("`chain_id`=?", stats.ChainId).Update("total_recycle", stats.TotalRecycle+recycleSNFT).Error
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
	stats.TotalBlock++
	stats.TotalTransaction += int64(len(parsed.CacheTxs))
	stats.TotalInternalTx += int64(len(parsed.CacheInternalTxs))
	stats.TotalNFT += int64(len(parsed.NFTs))
	stats.RewardSNFTCount += rewardSNFT
	stats.RewardCoinCount += int64(len(parsed.Rewards)) - rewardSNFT
	stats.TotalAccount = int64(len(stats.Balances))
	stats.TotalRecycle += recycleSNFT
	stats.TotalBalance = totalBalance.Text(10)
	stats.TotalAmount = totalAmount.Text(10)
	stats.TotalPledge = totalPledge.Text(10)
	stats.TotalNFTTx = totalNFTTx
	stats.TotalSNFTTx = totalSNFTTx
	stats.TotalStakerTx = totalStakerTx
	stats.TotalNFTAmount = totalNFTAmount.Text(10)
	stats.TotalSNFTAmount = totalSNFTAmount.Text(10)
	stats.TotalEpoch = stats.RewardSNFTCount/4096 + 1
	if parsed.Number > 0 && parsed.Miner == types.ZeroAddress {
		stats.TotalBlackHole++
	}
	for _, tx := range parsed.CacheTxs {
		if tx.Input == "0x" {
			stats.TotalTransferTx++
		} else if len(tx.Input) > 14 && tx.Input[:14] == "0x65726269653a" {
			stats.TotalErbieTx++
		}
	}
	return
}

func fixStats(db *gorm.DB, parsed *model.Parsed) (err error) {
	for _, account := range parsed.CacheAccounts {
		if account.Balance == "0x0" && account.Nonce == 0 && account.SNFTValue == "0" {
			if err = db.Delete(&model.Account{}, "`address`=?", account.Address).Error; err != nil {
				return
			}
			delete(stats.Balances, account.Address)
		} else {
			if err = db.Select("balance", "nonce", "number", "snft_value").Updates(account).Error; err != nil {
				return
			}
			if err = db.Model(&model.Account{}).Where("address=?", account.Address).Update(
				"snft_count",
				db.Model(&model.SNFT{}).Where("owner=? AND remove=false", account.Address).Select("COUNT(*)"),
			).Error; err != nil {
				return
			}
			stats.Balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
		}
	}
	return loadStats(db)
}

func freshStats(db *gorm.DB, parsed *model.Parsed) {
	if stats.Ready {
		for _, account := range parsed.CacheAccounts {
			db.Model(&model.Account{}).Where("address=?", account.Address).Update("snft_count", db.Model(&model.SNFT{}).Where("owner=?", account.Address).Select("count(*)"))
		}
		if number := parsed.Number; stats.TotalValidator == 0 || number%24 == 0 {
			if number > 1000 {
				db.Raw("SELECT (SELECT timestamp FROM blocks WHERE number=?)-(SELECT timestamp FROM blocks WHERE number=?)", number, number-1000).Scan(&stats.AvgBlockTime)
			}
			db.Model(&model.Creator{}).Count(&stats.TotalCreator)
			db.Model(&model.SNFT{}).Where("remove=false").Count(&stats.TotalSNFT)
			db.Model(&model.Staker{}).Count(&stats.TotalStaker)
			db.Model(&model.Validator{}).Where("weight>0").Count(&stats.TotalValidator)
			db.Model(&model.NFT{}).Select("COUNT(DISTINCT creator)").Scan(&stats.TotalNFTCreator)
			db.Model(&model.Epoch{}).Select("COUNT(DISTINCT creator)").Scan(&stats.TotalSNFTCreator)
			db.Model(&model.Validator{}).Where("weight>=10").Count(&stats.TotalValidatorOnline)
			db.Model(&model.Transaction{}).Where("block_number>?", parsed.Number-10000).Select("COUNT(DISTINCT `from`)").Scan(&stats.ActiveAccount)
			if stats.Total24HTx == 0 || number%720 == 0 {
				start, stop := utils.LastTimeRange(1)
				db.Model(&model.Transaction{}).Joins("LEFT JOIN blocks ON block_hash = blocks.hash").Where("timestamp>=? AND timestamp<?", start, stop).Count(&stats.Total24HTx)
				db.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL AND timestamp>=? AND timestamp<?", start, stop).Count(&stats.Total24HStakerTx)
				db.Model(&model.NFT{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&stats.Total24HNFT)

				var creators []*struct {
					Reward string
					Profit string
				}
				totalProfit, value := new(big.Int), new(big.Int)
				db.Model(&model.Creator{}).Find(&creators)
				for _, creator := range creators {
					value.SetString(creator.Profit, 10)
					totalProfit = totalProfit.Add(totalProfit, value)
					value.SetString(creator.Reward, 10)
					totalProfit = totalProfit.Add(totalProfit, value)
				}
				stats.TotalProfit = totalProfit.Text(10)
			}
			var validators []*model.Validator
			db.Where("weight>0").Find(&validators)
			for _, validator := range validators {
				switch validator.Weight {
				case 70:
					validator.Score = 50
				case 50:
					validator.Score = 40
				case 30, 10:
					validator.Score = 0
				default:
					validator.Score = -50
				}
				score := validator.RewardCount * stats.TotalValidator * 20 / stats.RewardCoinCount
				if score > 20 {
					validator.Score += 20
				} else {
					validator.Score += score
				}
				score = 30 - (stats.TotalBlock-validator.RewardNumber)/stats.TotalValidator
				if score > 0 {
					validator.Score += score
				}
				db.Select("score").Updates(validator)
			}
		}
	}
}

func GetStats() *model.Stats {
	return stats
}
