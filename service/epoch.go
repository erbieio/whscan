package service

import "server/common/model"

// EpochsRes system NFT period paging return parameters
type EpochsRes struct {
	Total  int64         `json:"total"`  //The total number of NFT periods in the system
	Epochs []model.Epoch `json:"epochs"` //List of system NFT periods
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
	Collections []model.Collection `json:"collections" gorm:"-"` // 16 collection information
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
