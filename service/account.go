package service

import (
	"server/common/model"
	"server/common/types"
)

// AccountsRes account paging return parameters
type AccountsRes struct {
	Total    int64           `json:"total"`    //Total number of accounts
	Balance  types.BigInt    `json:"balance"`  //The total amount of coins in the chain
	Accounts []model.Account `json:"accounts"` //Account list
}

func FetchAccounts(page, size int) (res AccountsRes, err error) {
	err = DB.Order("length(balance) DESC,balance DESC").Offset((page - 1) * size).Limit(size).Find(&res.Accounts).Error
	// use cache to speed up queries
	res.Balance = TotalBalance()
	res.Total = int64(cache.TotalAccount)
	return
}

type AccountRes struct {
	model.Account
	PledgeAmount    string `json:"pledgeAmount"`    // pledge amount
	TotalNFT        uint64 `json:"totalNFT"`        // Number of NFTs held
	TotalSNFT       uint64 `json:"totalSNFT"`       // Number of SNFTs held
	RewardCoinCount uint64 `json:"rewardCoinCount"` // Number of times to get coin rewards, 0.1ERB once
	RewardSNFTCount uint64 `json:"rewardSNFTCount"` // Number of times to get SNFT rewards
}

func GetAccount(addr string) (res AccountRes, err error) {
	s := "*, (SELECT COUNT(*) FROM nfts WHERE owner=accounts.address) AS total_nft"
	s += ", (SELECT COUNT(*) FROM snfts WHERE owner=accounts.address) AS total_snft"
	s += ", IFNULL((SELECT amount FROM consensus_pledges WHERE address=accounts.address),'0') AS pledge_amount"
	s += ", (SELECT COUNT(amount) FROM rewards WHERE address=accounts.address) AS reward_coin_count"
	s += ", (SELECT COUNT(snft) FROM rewards WHERE address=accounts.address) AS reward_snft_count"
	err = DB.Model(model.Account{}).Where("address=?", addr).Select(s).Scan(&res).Error
	return
}