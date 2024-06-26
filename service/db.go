package service

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"server/common/model"
	. "server/conf"
)

var DB *gorm.DB

func init() {
	var err error
	DB, err = gorm.Open(mysql.Open(MysqlDsn+"?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		panic(err)
	}
	// Synchronize the table structure to the database, compare the structure in the database and the code, and perform DDL operations
	if err = model.Migrate(DB); err != nil {
		panic(err)
	}
	if err = initStats(DB); err != nil {
		panic(err)
	}
	if err = initValidator(DB); err != nil {
		panic(err)
	}
}
