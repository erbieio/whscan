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
		&Exchanger{},
		&UserNFT{},
		&NFTTx{},
		&ConsensusPledge{},
		&ExchangerPledge{},
		&SNFT{},
	).Error
}

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	ERBPay   string `json:"ERBPay"`
	ResetDB  bool   `json:"reset_db"`
}

var Conf Config

func NewMysqlDb() {
	filePtr, err := os.Open("./config.json")
	if err != nil {
		log.Info("InitEnv failed!")
	}
	defer filePtr.Close()
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&Conf)

	DB, err = gorm.Open("mysql", Conf.User+":"+Conf.Password+"@tcp("+Conf.Host+":"+Conf.Port+")/"+Conf.Database+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

	// 重置数据库表
	if Conf.ResetDB {
		err = DB.DropTableIfExists(
			&BlockModel{},
			&TransactionModel{},
			&TxLogModel{},
			&Exchanger{},
			&UserNFT{},
			&NFTTx{},
			&ConsensusPledge{},
			&ExchangerPledge{},
			&SNFT{},
		).Error
		if err != nil {
			panic(err)
		}
	}
}
