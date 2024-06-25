package service

import (
	"server/common/model"
)

// AccountsRes account paging return parameters
type AccountsRes struct {
	Total    int64  `json:"total"`   //Total number of accounts
	Balance  string `json:"balance"` //The total amount of coins in the chain
	Accounts []*struct {
		model.Account
		ValidatorAmount string `json:"validatorAmount,omitempty"` // validator pledge amount
		StakerAmount    string `json:"stakerAmount,omitempty"`    // staker pledge amount
	} `json:"accounts"` //Account list
}

func FetchAccounts(page, size int, order string) (res AccountsRes, err error) {
	db := DB.Model(&model.Account{}).Joins("LEFT JOIN validators ON validators.address=accounts.address").
		Joins("LEFT JOIN stakers ON stakers.address=accounts.address")
	if order != "" {
		db = db.Order(order)
	}
	err = db.Offset((page - 1) * size).Limit(size).
		Select("accounts.*,validators.amount AS validator_amount,stakers.amount AS staker_amount").Scan(&res.Accounts).Error
	// use stats to speed up queries
	res.Balance = stats.TotalBalance
	res.Total = stats.TotalAccount
	return
}

type AccountRes struct {
	model.Account
	Weight          int64  `json:"weight"`          //online weight,if it is not 70, it means that it is not online
	NFTCount        int64  `json:"nftCount"`        // hold NFT number
	ValidatorAmount string `json:"validatorAmount"` // validator pledge amount
	StakerAmount    string `json:"stakerAmount"`    // staker pledge amount
	RewardCoinCount int64  `json:"rewardCoinCount"` // BlockNumber of times to get coin rewards, 0.1ERB once
	RewardSNFTCount int64  `json:"rewardSNFTCount"` // BlockNumber of times to get SNFT rewards
	ValidatorReward string `json:"validatorReward"` // validator reward
	StakerReward    string `json:"stakerReward"`    // staker reward
	LastNumber      int64  `json:"lastNumber"`
	Reward          string `json:"reward"` //vote profit
	Profit          string `json:"profit"` //royalty profit
}

func GetAccount(addr string) (res AccountRes, err error) {
	db := DB.Model(model.Account{}).Joins("LEFT JOIN validators ON validators.address=accounts.address")
	db = db.Joins("LEFT JOIN stakers ON stakers.address=accounts.address")
	db = db.Joins("LEFT JOIN creators ON creators.address=accounts.address")
	s := "accounts.*, creators.last_number, IFNULL(creators.reward,'0') AS reward, (SELECT COUNT(*) FROM nfts WHERE owner=accounts.address) AS nft_count"
	s += ", validators.weight AS weight, IFNULL(validators.amount,'0') AS validator_amount, validators.reward_count AS reward_coin_count"
	s += ", IFNULL(validators.reward,'0') AS validator_reward, IFNULL(stakers.amount,'0') AS staker_amount, IFNULL(profit,'0') AS profit, " +
		"stakers.reward_count AS reward_snft_count, IFNULL(stakers.reward,'0') AS staker_reward"
	err = db.Select(s).Where("accounts.address=?", addr).Scan(&res).Error
	return
}
