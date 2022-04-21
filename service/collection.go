package service

import "server/model"

// CollectionsRes 合集分页返回参数
type CollectionsRes struct {
	Total       int64              `json:"total"`       //NFT合集总数
	Collections []model.Collection `json:"collections"` //NFT合集列表
}

func FetchCollections(exchanger, creator string, page, size int) (res CollectionsRes, err error) {
	if exchanger != "" || creator != "" {
		where := ""
		if exchanger != "" {
			where += "exchanger='" + exchanger + "'"
		}
		if creator != "" {
			if exchanger != "" {
				where += " AND "
			}
			where += "creator='" + creator + "'"
		}
		err = DB.Where(where).Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Collections).Error
		if err != nil {
			return
		}
		err = DB.Where(where).Model(&model.Collection{}).Count(&res.Total).Error
	} else {
		err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Collections).Error
		if err != nil {
			return
		}
		err = DB.Model(&model.Collection{}).Count(&res.Total).Error
	}
	return
}

func GetCollection(id string) (c model.Collection, err error) {
	err = DB.Where("id=?", id).First(&c).Error
	return
}
