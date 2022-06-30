package service

import (
	"time"

	"server/common/model"
	"server/common/types"
)

// ExchangersRes exchange paging return parameters
type ExchangersRes struct {
	Total      int64             `json:"total"`        //Total number of exchanges
	Last0Total int64             `json:"last_0_total"` //The number of newly opened exchanges in the latest 0 days (today), calculated in real time
	Last1Total int64             `json:"last_1_total"` //Number of newly opened exchanges in the latest 1 day (yesterday), cached
	Last7Total int64             `json:"last_7_total"` //Number of newly opened exchanges in the last 7 days, cached
	Exchangers []model.Exchanger `json:"exchangers"`   //List of exchanges
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
		res.Total = int64(cache.TotalExchanger)
	}
	if err == nil {
		res.Last0Total, err = TodayExchangerTotal()
		if err != nil {
			return
		}
		last1Total, last7Total, err = cacheLastTotal()
		if err != nil {
			return
		}
		res.Last1Total = last1Total
		res.Last7Total = last7Total
	}
	return
}

// TodayExchangerTotal number of new exchanges opened today
func TodayExchangerTotal() (count int64, err error) {
	now := time.Now().Local()
	loc, _ := time.LoadLocation("Local")
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	start := startTime.Unix()
	err = DB.Model(&model.Exchanger{}).Where("timestamp>=? AND timestamp<?", start, now.Unix()).Count(&count).Error
	return
}

var lastExchangerTime string
var last1Total int64
var last7Total int64

func cacheLastTotal() (total1 int64, total2 int64, err error) {
	if now := time.Now().Local().Format("2006-01-02"); now != lastExchangerTime {
		last1Total, err = lastExchangerTotal(1)
		if err != nil {
			return
		}
		last7Total, err = lastExchangerTotal(7)
		if err != nil {
			return
		}
		lastExchangerTime = now
	}
	return last1Total, last7Total, nil
}

// lastExchangerTotal the latest number of newly opened exchanges
func lastExchangerTotal(day int64) (count int64, err error) {
	now := time.Now().Local()
	loc, _ := time.LoadLocation("Local")
	daySecond := 24 * time.Hour.Milliseconds() / 1000
	stopTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	stop := stopTime.Unix()
	start := stop - daySecond*day
	err = DB.Model(&model.Exchanger{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&count).Error
	return
}

func FindExchanger(addr types.Address) (res model.Exchanger, err error) {
	err = DB.Where("address=?", addr).First(&res).Error
	if err != nil {
		return
	}
	err = DB.Where("exchanger=?", addr).Model(&model.Collection{}).Select("COUNT(*)").Scan(&res.CollectionCount).Error
	return
}
