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
	// use stats to speed up queries
	res.Total = int64(TotalBlock())
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}

func FetchTotals() (Stats, error) {
	return stats, nil
}

func FetchValidator(page, size int) (res []*model.Validator, err error) {
	err = DB.Offset((page - 1) * size).Limit(size).Find(&res).Error
	return
}

func FetchLocations() (res []*model.Location, err error) {
	err = DB.Where("`address` IN (?)", DB.Model(&model.Validator{}).Select("proxy")).Find(&res).Error
	return
}
