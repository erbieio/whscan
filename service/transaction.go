package service

import (
	"encoding/hex"
	"server/common/model"
	"server/common/types"
	"strings"
	"time"
)

func GetTransaction(hash string) (res model.Transaction, err error) {
	err = DB.Where("transactions.hash=?", hash).Take(&res).Error
	return
}

// TransactionsRes transaction paging return parameters
type TransactionsRes struct {
	Total        int64                `json:"total"`        //The total number of transactions
	Transactions []*model.Transaction `json:"transactions"` //transaction list
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
		db.Joins("LEFT JOIN erbies ON hash=tx_hash").Where("`type` IN (?)", strings.Split(types, ","))
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
	err = db.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Transactions).Error
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

func GetInternalTransaction(hash string) (res []*model.InternalTx, err error) {
	err = DB.Where("`tx_hash`=?", hash).Find(&res).Error
	return
}

// ErbiesRes erbie transaction paging return parameters
type ErbiesRes struct {
	Total int64          `json:"total"` //The total number of transactions
	Data  []*model.Erbie `json:"data"`  //erbie transaction list
}

func FetchErbieTxs(page, size int, number, address, epoch, account, types string) (res ErbiesRes, err error) {
	db := DB.Model(&model.Erbie{})
	if number != "" {
		db = db.Where("`block_number`=?", number)
	}
	if address != "" {
		db = db.Where("`address`=?", address)
	}
	if epoch != "" {
		db = db.Where("LEFT(address,39)=?", epoch)
	}
	if account != "" {
		db = db.Where("`from`=? OR `to`=?", account, account)
	}
	if types != "" {
		db = db.Where("`type` IN (?)", strings.Split(types, ","))
	}

	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}

func GetErbieTransaction(hash string) (res model.Erbie, err error) {
	err = DB.Where("`tx_hash`=?", hash).Take(&res).Error
	return
}

func GetTransactionGrowthRate() (float64, error) {
	var TotalNum int64
	err := DB.Model(&model.Transaction{}).Count(&TotalNum).Error
	if err != nil {
		return 0.0, err
	}

	var TotalNum24 int64
	err = DB.Model(&model.Transaction{}).Where("timestamp >= ?", time.Now().Unix()-86400).Count(&TotalNum24).Error
	if err != nil {
		return 0.0, err
	}

	rate := float64(TotalNum24) / float64(TotalNum)

	return rate, nil
}

func FetchTransactionsOfAddress(page, size int, number, addr string) (res TransactionsRes, err error) {
	db := DB.Model(&model.Transaction{})
	if number != "" {
		db = db.Where("block_number=?", number)
	}
	if addr != "" {
		db = db.Where("`from`=? OR `to`=?", addr, addr)
	}
	if number != "" || addr != "" {
		err = db.Count(&res.Total).Error
	} else {
		// use stats to speed up queries
		res.Total = stats.TotalTransaction
	}
	if err != nil {
		return
	}
	err = db.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Transactions).Error

	// 跟新合约交易的交易金额
	// Contract tx
	contractHashs := make([]types.Hash, 0)
	for _, tx := range res.Transactions {
		if len(tx.Input) > 2 {
			input, _ := hex.DecodeString(tx.Input[2:])
			if len(input) >= 6 && string(input[0:6]) == "erbie:" {
				// 是erbie交易
				continue
			}
			contractHashs = append(contractHashs, tx.Hash)
		}
	}
	var contractTxs []model.ContractTx
	if len(contractHashs) > 0 {
		err = DB.Model(&model.ContractTx{}).Where("hash in ?", contractHashs).Find(&contractTxs).Error
		if err != nil {
			return
		}
	}
	for idx, tx := range res.Transactions {
		for _, cTx := range contractTxs {
			if tx.Hash == cTx.Hash {
				res.Transactions[idx].Value = cTx.Value
				break
			}
		}
	}

	return
}
