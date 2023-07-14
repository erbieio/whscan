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
	res.Total = stats.TotalBlock
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}

// PledgesRes pledge paging return parameters
type PledgesRes struct {
	Total int64          `json:"total"` //The total number of pledges
	Data  []model.Pledge `json:"data"`  //pledge list
}

func FetchPledges(staker, validator string, page, size int) (res PledgesRes, err error) {
	db := DB.Model(&model.Pledge{})
	if staker != "" {
		db = db.Where("staker=?", staker)
	}
	if validator != "" {
		db = db.Where("staker=?", validator)
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	if err != nil {
		return
	}
	err = db.Count(&res.Total).Error
	return
}
