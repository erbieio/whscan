package service

import "server/common/model"

// TransactionsRes transaction paging return parameters
type TransactionsRes struct {
	Total        int64               `json:"total"`        //The total number of transactions
	Transactions []model.Transaction `json:"transactions"` //Transaction list
}

func FetchTransactions(page, size int, number, addr *string) (res TransactionsRes, err error) {
	db := DB
	if number != nil {
		db = db.Where("block_number=?", *number)
	}
	if addr != nil {
		db = db.Where("`from`=? OR `to`=?", *addr, *addr)
	}
	if number != nil || addr != nil {
		err = db.Model(&model.Transaction{}).Count(&res.Total).Error
	} else {
		// use cache to speed up queries
		res.Total = int64(TotalTransaction())
	}
	if err != nil {
		return
	}
	err = db.Order("block_number DESC, tx_index DESC").Offset((page - 1) * size).Limit(size).Find(&res.Transactions).Error
	return
}

func GetTransaction(hash string) (t model.Transaction, err error) {
	err = DB.Where("hash=?", hash).First(&t).Error
	return
}

func GetTransactionLogs(hash string) (t []model.Log, err error) {
	err = DB.Where("tx_hash=?", hash).Find(&t).Error
	return
}
