package service

import (
	"server/common/model"
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

func FetchTotals() (res Cache, err error) {
	return cache, nil
}
