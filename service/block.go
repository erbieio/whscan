package service

import "server/model"

// BlocksRes 区块分页返回参数
type BlocksRes struct {
	Total  int64         `json:"total"`  //区块总数
	Blocks []model.Block `json:"blocks"` //区块列表
}

func FetchBlocks(page, size int) (res BlocksRes, err error) {
	err = DB.Order("number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Blocks).Error
	// 使用缓存加速查询
	res.Total = int64(TotalBlock())
	return
}

func GetBlock(number string) (b model.Block, err error) {
	err = DB.Where("number=?", number).First(&b).Error
	return
}
