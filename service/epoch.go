package service

import "server/common/model"

// EpochsRes system NFT period paging return parameters
type EpochsRes struct {
	Total  int64         `json:"total"`  //The total number of NFT periods in the system
	Epochs []model.Epoch `json:"epochs"` //List of system NFT periods
}

func FetchEpochs(creator, order string, page, size int) (res EpochsRes, err error) {
	db := DB.Model(&Epoch{})
	if creator != "" {
		db = db.Where("creator=?", creator)
	}
	if order != "" {
		db = db.Order(order)
	} else {
		db = db.Order("id DESC")
	}
	if creator == "" {
		res.Total = stats.TotalEpoch
	} else {
		if err = db.Count(&res.Total).Error; err != nil {
			return
		}
	}
	err = db.Offset((page - 1) * size).Limit(size).Scan(&res.Epochs).Error
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
	err = DB.Where("LEFT(id,39)=?", res.ID).Find(&res.Collections).Error
	return
}
