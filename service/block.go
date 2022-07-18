package service

import (
	"server/common/model"
)

// BlocksRes block paging return parameters
type BlocksRes struct {
	Total  int64 `json:"total"` //The total number of blocks
	Blocks []*struct {
		model.Block
		CoinCount uint64 `json:"coinCount"` //number of times to get coin rewards, 0.1ERB once
		SNFTCount uint64 `json:"snftCount"` //number of times to get SNFT rewards
	} `json:"blocks"` //block list
}

func FetchBlocks(page, size int) (res BlocksRes, err error) {
	err = DB.Model(&model.Block{}).Joins("RIGHT JOIN rewards ON block_number=number").
		Group("number").Select("blocks.*, COUNT(amount) AS coin_count, COUNT(snft) AS snft_count").
		Order("number DESC").Offset((page - 1) * size).Limit(size).Scan(&res.Blocks).Error
	// use cache to speed up queries
	res.Total = int64(TotalBlock())
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}

func FetchTotals() (Cache, error) {
	return cache, nil
}
