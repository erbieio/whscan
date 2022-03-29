package database

import "github.com/jinzhu/gorm"

type BlockModel struct {
	gorm.Model
	Block
}
type Block struct {
	Number      string `gorm:"index" json:"number"           gencodec:"required"`
	ParentHash  string `json:"parentHash"       gencodec:"required"`
	UncleHash   string `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    string `json:"miner"            gencodec:"required"`
	Ts          string `json:"timestamp" `
	TxCount     int    `json:"tx_count"`
	Root        string `json:"stateRoot"        gencodec:"required"`
	TxHash      string `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash string `json:"receiptsRoot"     gencodec:"required"`
	GasLimit    string `json:"gasLimit"         gencodec:"required"`
	GasUsed     string `json:"gasUsed"          gencodec:"required"`
	Difficulty  string `json:"difficulty"       gencodec:"required"`
	Size        string `json:"size"`
	MixDigest   string `json:"mixHash"`
	Extra       string `gorm:"type:text" json:"extraData"        gencodec:"required"`
}

func (b Block) Insert() error {
	var m BlockModel
	m.Block = b
	return DB.Create(&m).Error
}

func FetchBlocks(page, size int) (data []Block, count int64, err error) {
	err = DB.Table("block_models").Count(&count).Error
	err = DB.Table("block_models").Limit(size).Offset((page - 1) * size).Order("id desc").Find(&data).Error
	return data, count, err
}

func FindBlock(num string) (data Block, err error) {
	db := DB.Debug().Table("block_models").Where("number=?", num)
	err = db.First(&data).Error
	return data, err
}

func BlockCount() (count uint64, err error) {
	err = DB.Model(&BlockModel{}).Count(&count).Error
	return
}
