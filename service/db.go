package service

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "server/conf"
	"server/model"
)

var DB *gorm.DB

func init() {
	var err error
	DB, err = gorm.Open(mysql.Open(MysqlDsn+"?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{PrepareStmt: true})
	if err != nil {
		panic(err)
	}
	if ResetDB {
		// 重置数据库
		err = model.DropTable(DB)
		if err != nil {
			panic(err)
		}
	}
	// 同步表结构到数据库, 对比数据库和代码中的结构，并执行DDL操作
	err = model.Migrate(DB)
	if err != nil {
		panic(err)
	}
	err = initCache()
	if err != nil {
		panic(err)
	}
}
