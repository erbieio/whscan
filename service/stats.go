package service

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm"
	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// Stats caches some database queries to speed up queries
type Stats struct {
	ChainId             int64        `json:"chainId"`             //chain id
	GenesisBalance      string       `json:"genesisBalance"`      //Total amount of coins created
	TotalBlock          uint64       `json:"totalBlock"`          //Total number of blocks
	TotalTransaction    uint64       `json:"totalTransaction"`    //Total number of transactions
	TotalInternalTx     uint64       `json:"totalInternalTx"`     //Total number of internal transactions
	TotalTransferTx     uint64       `json:"totalTransferTx"`     //Total number of  transfer transactions
	TotalWormholesTx    uint64       `json:"totalWormholesTx"`    //Total number of  wormholes transactions
	TotalUncle          uint64       `json:"totalUncle"`          //Number of total uncle blocks
	TotalAccount        uint64       `json:"totalAccount"`        //Total account number
	TotalBalance        string       `json:"totalBalance"`        //The total amount of coins in the chain
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
	TotalValidator      uint64       `json:"totalValidator"`      //Total number of validator
	TotalPledge         types.BigInt `json:"totalPledge"`         //Total amount of validator pledge

	genesis  model.Header
	balances map[types.Address]*big.Int
	fnfts    map[string]int64
}

var stats = Stats{
	TotalAmount:     "0",
	TotalNFTAmount:  "0",
	TotalSNFTAmount: "0",
	TotalPledge:     "0",
	balances:        make(map[types.Address]*big.Int),
	fnfts:           make(map[string]int64),
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
	if err = db.Model(&model.Block{}).Select("COUNT(*)").Scan(&stats.TotalBlock).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Select("COUNT(*)").Scan(&stats.TotalTransaction).Error; err != nil {
		return
	}
	if err = db.Model(&model.InternalTx{}).Select("COUNT(*)").Scan(&stats.TotalInternalTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Select("COUNT(*)").Where("input='0x'").Scan(&stats.TotalTransferTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Transaction{}).Select("COUNT(*)").Where("LEFT(input,22)='0x776f726d686f6c65733a'").Scan(&stats.TotalWormholesTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Uncle{}).Select("COUNT(*)").Scan(&stats.TotalUncle).Error; err != nil {
		return
	}
	if err = db.Model(&model.Exchanger{}).Select("COUNT(*)").Scan(&stats.TotalExchanger).Error; err != nil {
		return
	}
	if err = db.Model(&model.Collection{}).Where("length(id)!=40").Select("COUNT(*)").Scan(&stats.TotalNFTCollection).Error; err != nil {
		return
	}
	if err = db.Model(&model.Collection{}).Where("length(id)=40").Select("COUNT(*)").Scan(&stats.TotalSNFTCollection).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFT{}).Select("COUNT(*)").Scan(&stats.TotalNFT).Error; err != nil {
		return
	}
	if err = db.Model(&model.SNFT{}).Select("COUNT(*)").Scan(&stats.TotalSNFT).Error; err != nil {
		return
	}
	if err = db.Model(&model.Reward{}).Select("COUNT(snft)").Scan(&stats.RewardSNFTCount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Reward{}).Select("COUNT(amount)").Scan(&stats.RewardCoinCount).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x0'").Select("COUNT(*)").Scan(&stats.TotalNFTTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.NFTTx{}).Where("LEFT(nft_addr,3)='0x8'").Select("COUNT(*)").Scan(&stats.TotalSNFTTx).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalAmount").Select("value").Scan(&stats.TotalAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalNFTAmount").Select("value").Scan(&stats.TotalNFTAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalSNFTAmount").Select("value").Scan(&stats.TotalSNFTAmount).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalRecycle").Select("value").Scan(&stats.TotalRecycle).Error; err != nil {
		return
	}
	if err = db.Model(&model.Cache{}).Where("`key`=?", "TotalPledge").Select("value").Scan(&stats.TotalPledge).Error; err != nil {
		return
	}
	if err = db.Model(&model.Block{}).Find(&stats.genesis, "number=0").Error; err != nil {
		return
	}
	stats.TotalAccount = uint64(len(stats.balances))
	totalBalance := new(big.Int)
	for _, balance := range stats.balances {
		totalBalance = totalBalance.Add(totalBalance, balance)
	}
	stats.TotalBalance = totalBalance.Text(10)
	return
}

var lastTime time.Time

func updateStats(db *gorm.DB, parsed *model.Parsed) {
	totalBalance, _ := new(big.Int).SetString(stats.TotalBalance, 0)
	for _, account := range parsed.CacheAccounts {
		balance, _ := new(big.Int).SetString(string(account.Balance), 0)
		totalBalance = totalBalance.Add(totalBalance, balance)
		if balance = stats.balances[account.Address]; balance != nil {
			totalBalance = totalBalance.Sub(totalBalance, balance)
		}
	}
	if parsed.Number == 0 {
		if err := db.Create(&model.Cache{Key: "ChainId", Value: fmt.Sprintf("%v", stats.ChainId)}).Error; err != nil {
			return
		}
		if err := db.Create(&model.Cache{Key: "GenesisBalance", Value: totalBalance.Text(10)}).Error; err != nil {
			return
		}
		stats.GenesisBalance = totalBalance.Text(10)
	}
	for _, account := range parsed.CacheAccounts {
		stats.balances[account.Address], _ = new(big.Int).SetString(string(account.Balance), 0)
	}
	stats.TotalAccount = uint64(len(stats.balances))
	stats.TotalBalance = totalBalance.Text(10)
	if now := time.Now(); now.Minute() != lastTime.Minute() {
		var number uint64
		if err := DB.Model(&model.Pledge{}).Select("COUNT(*)").Scan(&number).Error; err == nil {
			stats.TotalValidator = number
		}
		if err := DB.Model(&model.NFT{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			stats.TotalNFTCreator = number
		}
		if err := DB.Model(&model.Epoch{}).Select("COUNT(DISTINCT creator)").Scan(&number).Error; err == nil {
			stats.TotalSNFTCreator = number
		}

		if err := DB.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL").Select("COUNT(*)").Scan(&number).Error; err == nil {
			stats.TotalExchangerTx = number
		}
		if now.Hour() != lastTime.Hour() {
			start, stop := utils.LastTimeRange(1)
			if err := DB.Model(&model.Block{}).Where("timestamp>=? AND timestamp<?", start, stop).Select("IFNULL(SUM(total_transaction),0)").Scan(&number).Error; err == nil {
				stats.Total24HTx = number
			}
			if err := DB.Model(&model.NFTTx{}).Where("exchanger_addr IS NOT NULL AND timestamp>=? AND timestamp<?", start, stop).Select("COUNT(*)").Scan(&number).Error; err == nil {
				stats.Total24HExchangerTx = number
			}
			if err := DB.Model(&model.NFT{}).Where("timestamp>=? AND timestamp<?", start, stop).Select("COUNT(*)").Scan(&number).Error; err == nil {
				stats.Total24HNFT = number
			}
		}
		for fnft := range stats.fnfts {
			err := DB.Exec("CAll fresh_c_snft(?)", fnft).Error
			if err != nil {
				log.Println("fresh com-snft error:", err)
			}
		}
		stats.fnfts = make(map[string]int64)
		lastTime = now
	}
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

func CheckStats(chainId types.Uint64, genesis *model.Header) (err error) {
	if stats.TotalBlock == 0 {
		stats.ChainId = int64(chainId)
		stats.genesis = *genesis
	} else if stats.ChainId != int64(chainId) || stats.genesis != *genesis {
		err = fmt.Errorf("database and chain information do not match, chain ID: %v %v, genesis block: %v %v", stats.ChainId, chainId, stats.genesis, genesis)
	}
	return
}

func TotalBlock() types.Uint64 {
	return types.Uint64(stats.TotalBlock)
}
