package service

import "server/common/model"

// EpochsRes 系统NFT期分页返回参数
type EpochsRes struct {
	Total  int64         `json:"total"`  //系统NFT期总数
	Epochs []model.Epoch `json:"epochs"` //系统NFT期列表
}

func FetchEpochs(page, size int) (res EpochsRes, err error) {
	err = DB.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&res.Epochs).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.Epoch{}).Count(&res.Total).Error
	return
}

type Epoch struct {
	model.Epoch
	Collections []model.Collection `json:"collections" gorm:"-"` // 16个合集信息
}

func GetEpoch(id string) (res Epoch, err error) {
	if id != "current" {
		err = DB.Where("id=?", id).First(&res).Error
	} else {
		err = DB.Order("id DESC").First(&res).Error
	}
	if err != nil {
		return
	}
	err = DB.Where("LEFT(id,38)=?", res.ID).Find(&res.Collections).Error
	return
}
