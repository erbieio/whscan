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

func FetchValidator(page, size int) (res []*model.Validator, err error) {
	err = DB.Offset((page - 1) * size).Limit(size).Find(&res).Error
	return
}

type LocationRes struct {
	model.Location
	Online bool `json:"online"`
}

func FetchLocations() (res []*LocationRes, err error) {
	err = DB.Model(&model.Location{}).Joins("LEFT JOIN `validators` ON `locations`.`address`=`validators`.`address`").
		Where("`validators`.`address` IS NOT NULL").Select("locations.*,online").Scan(&res).Error
	return
}
