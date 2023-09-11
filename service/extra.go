package service

import (
	"database/sql"
	"gorm.io/gorm"
	"server/common/model"
)

func ExecSql(sqlStr string) (result []map[string]any, err error) {
	err = DB.Transaction(func(db *gorm.DB) error {
		return db.Raw(sqlStr).Scan(&result).Error
	}, &sql.TxOptions{ReadOnly: true})
	return
}

// SlashingsRes slashing paging return parameters
type SlashingsRes struct {
	Total int64            `json:"total"` //The total number of slashing
	Data  []model.Slashing `json:"data"`  //slashing list
}

func FetchSlashings(address, number, reason string, page, size int) (res SlashingsRes, err error) {
	db := DB.Model(&model.Slashing{})
	if address != "" {
		db = db.Where("address=?", address)
	}
	if number != "" {
		db = db.Where("block_number=?", number)
	}
	if reason != "" {
		db = db.Where("reason=?", reason)
	}
	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}
