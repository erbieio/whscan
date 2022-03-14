package database

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
	"server/log"
)

var DB *gorm.DB

func init() {
	NewMysqlDb()
	//同步表结构
	migrate(DB)
}

func migrate(db *gorm.DB) error {
	//同步表结构到数据库, 对比数据库和代码中的结构，并执行DDL操作，使数据库结构和代码保持一致
	return db.AutoMigrate(
		&BlockModel{},
		&TransactionModel{},
		&TxLogModel{},
	).Error
}

func NewMysqlDb() {
	type Config struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
	}
	var conf Config
	filePtr, err := os.Open("./config.json")
	if err != nil {
		log.Info("InitEnv failed!")
	}
	defer filePtr.Close()
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&conf)

	DB, err = gorm.Open("mysql", conf.User+":"+conf.Password+"@tcp("+conf.Host+":"+conf.Port+")/"+conf.Database+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
}
