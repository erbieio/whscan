package service

import (
	"server/common/model"
	"server/common/utils"
)

type LineChartRes struct {
	Blocks []*struct {
		Number           uint64 `json:"number"`
		TotalTransaction uint64 `json:"txCount"`
	} `json:"blocks"`
	Txs []*struct {
		Hash     string `json:"hash"`
		GasPrice uint64 `json:"gasPrice"`
	} `json:"txs"`
	Exchangers []*struct {
		Index uint64 `json:"index"`
		Day   string `json:"day"`
		Num   uint64 `json:"num"`
	} `json:"exchangers"`
}

func LineChart(limit int) (res LineChartRes, err error) {
	err = DB.Model(&model.Block{}).Order("number DESC").Limit(limit).Scan(&res.Blocks).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.Transaction{}).Order("block_number DESC, tx_index DESC").Limit(limit).Scan(&res.Txs).Error
	if err != nil {
		return
	}
	start, stop := utils.LastTimeRange(int64(limit))
	err = DB.Table("(?) A", DB.Model(&model.Exchanger{}).Select("(timestamp-1656950400) DIV 86400 AS `index`, FROM_UNIXTIME(timestamp,'%Y-%m-%d') AS `day`").
		Where("timestamp>=? AND timestamp<?", start, stop)).Group("`index`, `day`").Select("*, COUNT(*) AS num").Scan(&res.Exchangers).Error
	return
}
