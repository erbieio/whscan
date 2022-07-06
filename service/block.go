package service

import (
	"server/common/model"
	"server/common/types"
)

// BlocksRes block paging return parameters
type BlocksRes struct {
	Total  int64         `json:"total"`  //The total number of blocks
	Blocks []model.Block `json:"blocks"` //block list
}

func FetchBlocks(page, size int) (res BlocksRes, err error) {
	err = DB.Order("number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Blocks).Error
	// use cache to speed up queries
	res.Total = int64(TotalBlock())
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}

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
	freshCache()
	res.Total = int64(cache.TotalAccount)
	return
}

func FetchTotals() (res Cache, err error) {
	freshCache()
	return cache, nil
}
