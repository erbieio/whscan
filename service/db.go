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
	DB, err = gorm.Open(mysql.Open(MysqlDsn+"?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if ResetDB {
		// reset the database
		err = model.DropTable(DB)
		if err != nil {
			panic(err)
		}
	}
	// Synchronize the table structure to the database, compare the structure in the database and the code, and perform DDL operations
	err = model.Migrate(DB)
	if err != nil {
		panic(err)
	}
	err = model.SetView(DB)
	if err != nil {
		panic(err)
	}
	err = model.SetProcedure(DB)
	if err != nil {
		panic(err)
	}
	DB.Exec("INSERT IGNORE INTO stats (chain_id,genesis_balance,total_amount,total_nft_amount,total_snft_amount,total_recycle)VALUES (\n    (SELECT value FROM caches WHERE `key`='ChainId'),\n    (SELECT value FROM caches WHERE `key`='GenesisBalance'),\n    (SELECT value FROM caches WHERE `key`='TotalAmount'),\n    (SELECT value FROM caches WHERE `key`='TotalNFTAmount'),\n    (SELECT value FROM caches WHERE `key`='TotalSNFTAmount'),\n    (SELECT value FROM caches WHERE `key`='TotalRecycle')\n    );")
	err = initStats(DB)
	if err != nil {
		panic(err)
	}
	err = initValidator(DB)
}
