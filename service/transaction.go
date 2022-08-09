package service

import "server/common/model"

// TransactionRes transaction return parameters
type TransactionRes struct {
	model.Transaction
	Timestamp uint64 `json:"timestamp"` //The event stamp of the block it is in
}

func GetTransaction(hash string) (res TransactionRes, err error) {
	err = DB.Model(&Transaction{}).Joins("LEFT JOIN blocks ON number=block_number").Where("transactions.hash=?", hash).
		Select("transactions.*,timestamp").First(&res).Error
	return
}

// TransactionsRes transaction paging return parameters
type TransactionsRes struct {
	Total        int64             `json:"total"`        //The total number of transactions
	Transactions []*TransactionRes `json:"transactions"` //Transaction list
}

func FetchTransactions(page, size int, number, addr *string) (res TransactionsRes, err error) {
	db := DB.Model(&Transaction{})
	if number != nil {
		db = db.Where("block_number=?", *number)
	}
	if addr != nil {
		db = db.Where("`from`=? OR `to`=?", *addr, *addr)
	}
	if number != nil || addr != nil {
		err = db.Count(&res.Total).Error
	} else {
		// use cache to speed up queries
		res.Total = int64(cache.TotalTransaction)
	}
	if err != nil {
		return
	}
	err = db.Joins("LEFT JOIN blocks ON number=block_number").Select("transactions.*,timestamp").
		Order("block_number DESC, tx_index DESC").Offset((page - 1) * size).Limit(size).Scan(&res.Transactions).Error
	return
}

func GetTransactionLogs(hash string) (t []model.Log, err error) {
	err = DB.Where("tx_hash=?", hash).Find(&t).Error
	return
}
