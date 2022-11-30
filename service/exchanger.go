package service

import (
	"time"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
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
		err = DB.Where("amount!='0' AND name=?", name).Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Exchangers).Error
		if err != nil {
			return
		}
		err = DB.Where("amount!='0' AND name=?", name).Model(&model.Exchanger{}).Count(&res.Total).Error
	} else {
		err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Exchangers).Error
		res.Total = stats.TotalExchanger
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
	start, stop := utils.LastTimeRange(day)
	err = DB.Model(&model.Exchanger{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&count).Error
	return
}

type ExchangerRes struct {
	model.Exchanger
	TxCount         uint64 `json:"txCount"`         //transaction count
	CollectionCount uint64 `json:"collectionCount"` //collection count
}

func FindExchanger(addr types.Address) (res ExchangerRes, err error) {
	err = DB.Model(&model.Exchanger{}).Where("address=?", addr).Scan(&res).Error
	if err != nil {
		return
	}
	err = DB.Where("exchanger=?", addr).Model(&model.Collection{}).Select("COUNT(*)").Scan(&res.CollectionCount).Error
	if err != nil {
		return
	}
	err = DB.Where("exchanger_addr=?", addr).Model(&model.NFTTx{}).Select("COUNT(*)").Scan(&res.TxCount).Error
	return
}

func Exchangers(page, size int, order string) (res []ExchangerRes, err error) {
	s := "*, (SELECT COUNT(*) FROM collections WHERE exchanger=exchangers.address) AS collection_count"
	s += ", (SELECT COUNT(*) FROM nft_txes WHERE exchanger_addr=exchangers.address) AS tx_count"
	db := DB.Model(model.Exchanger{}).Where("amount!='0'").Select(s).Offset((page - 1) * size).Limit(size)
	if order != "" {
		db = db.Order(order)
	}
	err = db.Scan(&res).Error
	return
}

type ExchangerTxCountRes struct {
	Total uint64 `json:"total"`
	Day30 uint64 `json:"day30"`
	Day7  uint64 `json:"day_7"`
	Day1  uint64 `json:"day_1"`
}

func ExchangerTxCount(addr types.Address) (res ExchangerTxCountRes, err error) {
	start, stop := utils.LastTimeRange(30)
	db := DB.Model(&model.NFTTx{}).Where("exchanger_addr=?", addr).Select("COUNT(*)")
	err = db.Scan(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Where("timestamp>=? AND timestamp<?", start, stop).Scan(&res.Day30).Error
	if err != nil {
		return
	}
	err = db.Where("timestamp>=? AND timestamp<?", stop-7*utils.DaySecond, stop).Scan(&res.Day7).Error
	if err != nil {
		return
	}
	err = db.Where("timestamp>=? AND timestamp<?", stop-1*utils.DaySecond, stop).Scan(&res.Day1).Error
	return
}
