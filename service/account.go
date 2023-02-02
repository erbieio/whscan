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
		ExchangerAmount string `json:"exchangerAmount,omitempty"` // exchanger pledge amount
	} `json:"accounts"` //Account list
}

func FetchAccounts(page, size int, order string) (res AccountsRes, err error) {
	db := DB.Model(&model.Account{}).Joins("LEFT JOIN validators ON validators.address=accounts.address").
		Joins("LEFT JOIN exchangers ON exchangers.address=accounts.address")
	if order != "" {
		db = db.Order(order)
	}
	err = db.Offset((page - 1) * size).Limit(size).
		Select("accounts.*,validators.amount AS validator_amount,exchangers.amount AS exchanger_amount").Scan(&res.Accounts).Error
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
	ExchangerAmount string `json:"exchangerAmount"` // exchanger pledge amount
	RewardCoinCount int64  `json:"rewardCoinCount"` // Number of times to get coin rewards, 0.1ERB once
	RewardSNFTCount int64  `json:"rewardSNFTCount"` // Number of times to get SNFT rewards
}

func GetAccount(addr string) (res AccountRes, err error) {
	s := "*, (SELECT weight FROM validators WHERE address=accounts.address)"
	s += ", (SELECT COUNT(*) FROM nfts WHERE owner=accounts.address) AS nft_count"
	s += ", IFNULL((SELECT amount FROM validators WHERE address=accounts.address),'0') AS validator_amount"
	s += ", IFNULL((SELECT amount FROM exchangers WHERE address=accounts.address),'0') AS exchanger_amount"
	s += ", (SELECT COUNT(amount) FROM rewards WHERE address=accounts.address) AS reward_coin_count"
	s += ", (SELECT COUNT(snft) FROM rewards WHERE address=accounts.address) AS reward_snft_count"
	err = DB.Model(model.Account{}).Where("address=?", addr).Select(s).Scan(&res).Error
	return
}
