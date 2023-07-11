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
	Stakers []*struct {
		Index uint64 `json:"index"`
		Day   string `json:"day"`
		Num   uint64 `json:"num"`
	} `json:"stakers"`
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
	err = DB.Table("(?) A", DB.Model(&model.Staker{}).Select("(timestamp-?) DIV 86400 AS `index`, FROM_UNIXTIME(timestamp,'%Y-%m-%d') AS `day`", start).
		Where("timestamp>=? AND timestamp<?", start, stop)).Group("`index`, `day`").Select("*, COUNT(*) AS num").Scan(&res.Stakers).Error
	return
}

type TxChartRes struct {
	Hour uint64 `json:"hour"` // hour
	Num  uint64 `json:"num"`  // number of transaction
}

func TxChart() (res []*TxChartRes, err error) {
	start, stop := utils.LastTimeRange(int64(1))
	err = DB.Table("(?) A", DB.Model(&model.Block{}).Select("(timestamp-?) DIV 3600 AS `hour`,`total_transaction`", start).
		Where("timestamp>=? AND timestamp<?", start, stop)).Group("`hour`").Order("`hour`").Select("`hour`, SUM(total_transaction) AS num").Scan(&res).Error
	return
}

type NFTChartRes struct {
	Hour uint64 `json:"hour"` // hour
	Num  uint64 `json:"num"`  // number of nft
}

func NFTChart() (res []*NFTChartRes, err error) {
	start, stop := utils.LastTimeRange(int64(1))
	err = DB.Table("(?) A", DB.Model(&model.NFT{}).Select("(timestamp-?) DIV 3600 AS `hour`,`address`", start).
		Where("timestamp>=? AND timestamp<?", start, stop)).Group("`hour`").Order("`hour`").Select("`hour`, COUNT(address) AS num").Scan(&res).Error
	return
}
