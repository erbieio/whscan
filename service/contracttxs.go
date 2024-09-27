package service

import (
	"gorm.io/gorm"
	"server/common/model"
)

// ContractsRes contract paging return parameters
type ContractTxsRes struct {
	Total       int64               `json:"total"`        //The total number of contracttxs
	ContractTxs []*model.ContractTx `json:"contract_txs"` //contract transaction list
}

func FetchContractTxs(addr string, page, size int) (res ContractTxsRes, err error) {
	var db *gorm.DB

	db = DB.Model(&model.ContractTx{}).Where("contract_txes.to = ?", addr).Order("block_number DESC")
	err = DB.Model(&model.ContractTx{}).Where("contract_txes.to = ?", addr).Count(&res.Total).Error
	if err != nil {
		return
	}

	err = db.Offset((page - 1) * size).Limit(size).Find(&res.ContractTxs).Error
	if err != nil {
		return
	}

	return
}
