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
	Address   string  `json:"address"`   //account address
	Proxy     string  `json:"proxy"`     //proxy address
	Latitude  float64 `json:"latitude"`  //latitude
	Longitude float64 `json:"longitude"` //longitude
	Online    bool    `json:"online"`    //online status
}

func FetchLocations() (res []*LocationRes, err error) {
	err = DB.Model(&model.Validator{}).Joins("LEFT JOIN `locations` ON `validators`.`proxy`=`locations`.`address`").
		Where("`amount`!='0'").Select("`validators`.`address`,`proxy`,`latitude`,`longitude`,`online`").Scan(&res).Error
	return
}
