package service

import "server/common/model"

type LineChartRes struct {
	Blocks []struct {
		Number           uint64 `json:"number"`
		TotalTransaction uint64 `json:"txCount"`
	} `json:"blocks"`
	Txs []struct {
		Hash     string `json:"hash"`
		GasPrice uint64 `json:"gasPrice"`
	}
}

func LineChart(limit int) (res LineChartRes, err error) {
	err = DB.Model(&model.Block{}).Order("number DESC").Limit(limit).Scan(&res.Blocks).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.Transaction{}).Order("block_number DESC, tx_index DESC").Limit(limit).Scan(&res.Txs).Error
	return
}
