package service

import (
	"errors"
	"strings"

	"server/common/model"
)

// TransactionRes transaction return parameters
type TransactionRes struct {
	model.Transaction
	Timestamp uint64 `json:"timestamp"` //The event stamp of the block it is in
}

func GetTransaction(hash string) (res TransactionRes, err error) {
	err = DB.Model(&model.Transaction{}).Joins("LEFT JOIN blocks ON number=block_number").Where("transactions.hash=?", hash).
		Select("transactions.*,timestamp").First(&res).Error
	return
}

// TransactionsRes transaction paging return parameters
type TransactionsRes struct {
	Total        int64             `json:"total"`        //The total number of transactions
	Transactions []*TransactionRes `json:"transactions"` //transaction list
}

func FetchTransactions(page, size int, number, addr, types string) (res TransactionsRes, err error) {
	db := DB.Model(&model.Transaction{})
	if number != "" {
		db = db.Where("block_number=?", number)
	}
	if addr != "" {
		db = db.Where("`from`=? OR `to`=?", addr, addr)
	}
	if types != "" {
		conditions := ""
		for _, t := range strings.Split(types, ",") {
			if conditions != "" {
				conditions += " OR "
			}
			switch t {
			case "0":
				conditions += "LEFT(input,34)='0x65726269653a7b2274797065223a302c'"
			case "1":
				conditions += "LEFT(input,34)='0x65726269653a7b2274797065223a312c'"
			case "6":
				conditions += "LEFT(input,34)='0x65726269653a7b2274797065223a362c'"
			case "9":
				conditions += "LEFT(input,34)='0x65726269653a7b2274797065223a392c'"
			case "10":
				conditions += "LEFT(input,36)='0x65726269653a7b2274797065223a31302c'"
			case "26":
				conditions += "LEFT(input,36)='0x65726269653a7b2274797065223a32362c'"
			case "31":
				conditions += "LEFT(input,36)='0x65726269653a7b2274797065223a33312c'"
			default:
				return TransactionsRes{}, errors.New(t + " not support erbie tx type")
			}
		}
		db = db.Where(conditions)
	}
	if number != "" || addr != "" || types != "" {
		err = db.Count(&res.Total).Error
	} else {
		// use stats to speed up queries
		res.Total = stats.TotalTransaction
	}
	if err != nil {
		return
	}
	err = db.Joins("LEFT JOIN blocks ON number=block_number").Select("transactions.*,timestamp").
		Order("block_number DESC").Offset((page - 1) * size).Limit(size).Scan(&res.Transactions).Error
	return
}

func GetTransactionLogs(hash string) (t []model.EventLog, err error) {
	err = DB.Where("tx_hash=?", hash).Find(&t).Error
	return
}

// InternalTxsRes internal transaction paging return parameters
type InternalTxsRes struct {
	Total       int64               `json:"total"`        //The total number
	InternalTxs []*model.InternalTx `json:"internal_txs"` //transaction list
}

func GetInternalTransactions(page, size int) (res InternalTxsRes, err error) {
	err = DB.Order("`block_number` DESC").Offset((page - 1) * size).Limit(size).Find(&res.InternalTxs).Error
	res.Total = stats.TotalInternalTx
	return
}

func GetInternalTransaction(hash string) (t []*model.InternalTx, err error) {
	err = DB.Where("`tx_hash`=?", hash).Find(&t).Error
	return
}
