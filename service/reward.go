package service

import "server/common/model"

// RewardsRes reward paging return parameters
type RewardsRes struct {
	Total   int64 `json:"total"` //The total number of rewards
	Rewards []*struct {
		model.Reward
		CollectionName string `json:"collectionName"`
	} `json:"rewards"` //Rewards list
}

func FetchRewards(page, size int) (res RewardsRes, err error) {
	err = DB.Model(&model.Reward{}).Order("block_number DESC").Offset((page - 1) * size).Limit(size).
		Joins("LEFT JOIN collections ON id=LEFT(snft,40)").Select("rewards.*, name AS collection_name").Scan(&res.Rewards).Error
	res.Total = int64((stats.TotalBlock - 1) * 11)
	return
}

func BlockRewards(block string) (res []model.Reward, err error) {
	err = DB.Where("block_number=?", block).Find(&res).Error
	return
}
