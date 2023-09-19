package service

import "server/common/model"

// RewardsRes reward paging return parameters
type RewardsRes struct {
	Total   int64           `json:"total"`   //The total number of rewards
	Rewards []*model.Reward `json:"rewards"` //Rewards list
}

func FetchRewards(page, size int) (res RewardsRes, err error) {
	err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Rewards).Error
	res.Total = (stats.TotalBlock - stats.TotalBlackHole - 1) * 11
	return
}

func BlockRewards(block string) (res []*model.Reward, err error) {
	err = DB.Where("block_number=?", block).Find(&res).Error
	return
}
