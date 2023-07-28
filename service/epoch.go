package service

import "server/common/model"

// EpochsRes system NFT period paging return parameters
type EpochsRes struct {
	Total  int64         `json:"total"`  //The total number of NFT periods in the system
	Epochs []model.Epoch `json:"epochs"` //List of system NFT periods
}

func FetchEpochs(creator, order string, page, size int) (res EpochsRes, err error) {
	db := DB.Model(&model.Epoch{})
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

func GetEpoch(id string) (res model.Epoch, err error) {
	if id != "current" {
		err = DB.Where("id=?", id).First(&res).Error
	} else {
		err = DB.Order("number DESC").First(&res).Error
	}
	return
}
