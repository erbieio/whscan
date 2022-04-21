package model

import "gorm.io/gorm"

var Tables = []interface{}{
	&Block{},
	&Uncle{},
	&Transaction{},
	&Log{},
	&Account{},
	&Contract{},
	&InternalTx{},
	&ERC20Transfer{},
	&ERC721Transfer{},
	&ERC1155Transfer{},
	&Exchanger{},
	&UserNFT{},
	&NFTTx{},
	&ConsensusPledge{},
	&ExchangerPledge{},
	&OfficialNFT{},
	&Collection{},
	&NFTMeta{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Tables...)
}

func DropTable(db *gorm.DB) error {
	return db.Migrator().DropTable(Tables...)
}
