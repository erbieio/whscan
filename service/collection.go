package service

import "server/common/model"

// CollectionsRes collection paging return parameters
type CollectionsRes struct {
	Total       int64              `json:"total"`       //The total number of NFT collections
	Collections []model.Collection `json:"collections"` //NFT collection list
}

func FetchCollections(exchanger, creator, _type string, page, size int) (res CollectionsRes, err error) {
	db := DB
	if exchanger != "" {
		db = db.Where("exchanger=?", exchanger)
	}
	if creator != "" {
		db = db.Where("creator=?", creator)
	}
	if _type == "nft" {
		db = db.Where("length(id)=64")
	}
	if _type == "snft" {
		db = db.Where("length(id)!=64")
	}

	err = db.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Collections).Error
	if err != nil {
		return
	}
	err = db.Model(&model.Collection{}).Count(&res.Total).Error
	return
}

func GetCollection(id string) (c model.Collection, err error) {
	err = DB.Where("id=?", id).First(&c).Error
	return
}
