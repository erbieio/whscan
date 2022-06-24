package service

import (
	"server/common/model"
	"server/common/types"
)

// BlocksRes 区块分页返回参数
type BlocksRes struct {
	Total  int64         `json:"total"`  //区块总数
	Blocks []model.Block `json:"blocks"` //区块列表
}

func FetchBlocks(page, size int) (res BlocksRes, err error) {
	err = DB.Order("number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Blocks).Error
	// 使用缓存加速查询
	res.Total = int64(TotalBlock())
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}

// AccountsRes 账户分页返回参数
type AccountsRes struct {
	Total    int64           `json:"total"`                        //账户总数
	Balance  *types.BigInt   `json:"balance" swaggertype:"string"` //链的币总额
	Accounts []model.Account `json:"blocks"`                       //账户列表
}

func FetchAccounts(page, size int) (res AccountsRes, err error) {
	err = DB.Order("length(balance) DESC,balance DESC").Offset((page - 1) * size).Limit(size).Find(&res.Accounts).Error
	// 使用缓存加速查询
	res.Balance = TotalBalance()
	res.Total = int64(TotalAccount())
	return
}
