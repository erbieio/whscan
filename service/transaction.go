package service

import "server/common/model"

// TransactionsRes 交易分页返回参数
type TransactionsRes struct {
	Total        int64               `json:"total"`        //交易总数
	Transactions []model.Transaction `json:"transactions"` //交易列表
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
		// 使用缓存加速查询
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
