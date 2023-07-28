package service

import (
	"server/common/model"
	"server/common/types"
)

// StakersRes stakers paging return parameters
type StakersRes struct {
	Total int64          `json:"total"` //Total number of stakers
	Data  []model.Staker `json:"data"`  //List of stakers
}

func FetchStakers(order string, page, size int) (res StakersRes, err error) {
	db := DB.Model(model.Staker{})
	if order != "" {
		db = db.Order(order)
	}
	res.Total = stats.TotalStaker
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}

func FindStaker(addr types.Address) (res model.Staker, err error) {
	err = DB.Find(&res, "address=?", addr).Error
	return
}
