package service

import (
	"server/model"
	"time"
)

// ExchangersRes 交易所分页返回参数
type ExchangersRes struct {
	Total          int64             `json:"total"`           //交易所总数
	YesterdayTotal int64             `json:"yesterday_total"` //昨日新开交易所数量
	Exchangers     []model.Exchanger `json:"exchangers"`      //交易所列表
}

func FetchExchangers(name string, page, size int) (res ExchangersRes, err error) {
	if name != "" {
		err = DB.Where("name=?", name).Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Exchangers).Error
		if err != nil {
			return
		}
		err = DB.Where("name=?", name).Model(&model.Exchanger{}).Count(&res.Total).Error
	} else {
		err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Exchangers).Error
		if err != nil {
			return
		}
		err = DB.Model(&model.Exchanger{}).Count(&res.Total).Error
	}
	if err == nil {
		res.YesterdayTotal, err = yesterdayExchangerTotal()
	}
	return
}

// yesterdayExchangerTotal 昨日创建的交易所数量
func yesterdayExchangerTotal() (count int64, err error) {
	now := time.Now().Local()
	loc, _ := time.LoadLocation("Local")
	daySecond := 24 * time.Hour.Milliseconds() / 1000
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	stop := startTime.Unix()
	start := stop - daySecond
	err = DB.Model(&model.Exchanger{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&count).Error
	return
}

func FindExchanger(addr string) (res model.Exchanger, err error) {
	err = DB.Where("address=?", addr).First(&res).Error
	if err != nil {
		return
	}
	err = DB.Where("exchanger_addr=?", addr).Model(&model.UserNFT{}).Select("COUNT(*)").Scan(&res.NFTCount).Error
	return
}
