package service

import "server/common/model"

// CreatorsRes creator paging return parameters
type CreatorsRes struct {
	Total    int64            `json:"total"`    //The total number of creator
	Creators []*model.Creator `json:"creators"` //List of creator
}

func FetchCreators(page, size int, order string) (res CreatorsRes, err error) {
	db := DB.Order(order).Offset((page - 1) * size).Limit(size)
	if order != "" {
		db = db.Order(order)
	}
	err = db.Find(&res.Creators).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.Creator{}).Count(&res.Total).Error
	return
}

func GetCreator(addr string) (res model.Creator, err error) {
	err = DB.Where("address=?", addr).First(&res).Error
	return
}
