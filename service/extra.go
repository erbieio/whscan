package service

import (
	"database/sql"
	"gorm.io/gorm"
)

func ExecSql(sqlStr string) (result []map[string]any, err error) {
	err = DB.Transaction(func(db *gorm.DB) error {
		return db.Raw(sqlStr).Scan(&result).Error
	}, &sql.TxOptions{ReadOnly: true})
	return
}
