package service

import (
	"server/common/model"
)

// BlocksRes block paging return parameters
type BlocksRes struct {
	Total  int64         `json:"total"`  //The total number of blocks
	Blocks []model.Block `json:"blocks"` //block list
}

func FetchBlocks(page, size int, filter string) (res BlocksRes, err error) {
	db := DB.Model(&model.Block{})
	if filter == "1" {
		db = db.Where("number!=0 AND `miner` = '0x0000000000000000000000000000000000000000'")
	} else if filter == "2" {
		db = db.Where("`proof` != '[]'")
	}
	if filter != "" {
		if err = db.Count(&res.Total).Error; err != nil {
			return
		}
	} else {
		// use stats to speed up queries
		res.Total = stats.TotalBlock
	}

	err = db.Order("number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Blocks).Error
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
		db = db.Where("validator=?", validator)
	}
	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}
