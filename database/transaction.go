package database

import "github.com/jinzhu/gorm"

type TransactionModel struct {
	gorm.Model
	Transaction
}
type Transaction struct {
	Hash                  string `json:"hash"`
	BlockHash             string `json:"blockHash"`
	BlockNumber           string `gorm:"index" json:"blockNumber"`
	Ts                    string `gorm:"index" json:"timestamp" `
	From                  string `gorm:"index" json:"from"`
	To                    string `gorm:"index" json:"to"`
	InternalValueTransfer string `gorm:"type:text" json:"internal_value_transfer"`
	Value                 string `json:"value"`
	InternalCalls         string `gorm:"type:text" json:"internal_calls"`
	TokenTransfer         string `gorm:"type:text" json:"token_transfer"`
	Gas                   string `json:"gas"`
	GasPrice              string `json:"gasPrice"`
	TxType                string `gorm:"index" json:"tx_type"`
	TransactionIndex      string `json:"transactionIndex"`
	Nonce                 string `json:"nonce"`
	Input                 string `gorm:"type:text" json:"input"`
	Status                string `json:"status"`
}

func (t Transaction) Insert() error {
	var m TransactionModel
	m.Transaction=t
	return DB.Create(&m).Error
}

func FetchTxs(page, size int, addr, ty, block string) (data []Transaction, count int64, err error) {
	db := DB.Table("transaction_models")
	if addr != "" {
		db = db.Where("`from`=? or `to`=?", addr, addr)
	}
	if ty != "" {
		if ty == "internal" {
			db = db.Where("internal_calls!='[]'")
		} else {
			db = db.Where("tx_type like '%" + ty + "%'")
		}
	}
	if block != "" {
		db = db.Where("block_number=?", block)
	}
	err = db.Count(&count).Error
	err = db.Limit(size).Offset((page - 1) * size).Order("id desc").Find(&data).Error
	return data, count, err
}

func FindTx(txHash string) (data Transaction, err error) {
	err = DB.Debug().Table("transaction_models").Where("hash=?", txHash).First(&data).Error
	return data, err
}

type TxLogModel struct {
	gorm.Model
	TxLog
}
type TxLog struct {
	Address string `gorm:"index" json:"address"`
	Topics  string `gorm:"type:text" json:"topics"`
	Data    string `gorm:"type:text" json:"data"`
	TxHash  string `json:"tx_hash"`
}

func (t TxLog) Insert() error {
	var m TxLogModel
	m.TxLog=t
	return DB.Create(&m).Error
}

func FindLogByTx(txHash string) (data []TxLog, err error) {
	err = DB.Table("tx_log_models").Where("tx_hash=?", txHash).Order("id asc").Find(&data).Error
	return data, err
}
