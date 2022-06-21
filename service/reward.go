package service

import "server/common/model"

// RewardsRes 奖励分页返回参数
type RewardsRes struct {
	Total   int64          `json:"total"`   //奖励总数
	Rewards []model.Reward `json:"rewards"` //奖励列表
}

func FetchRewards(page, size int) (res RewardsRes, err error) {
	err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Rewards).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.Reward{}).Count(&res.Total).Error
	return
}

func BlockRewards(block string) (res []model.Reward, err error) {
	err = DB.Where("block_number=?", block).Find(&res).Error
	return
}
