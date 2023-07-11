package service

import (
	"time"

	"server/common/model"
	"server/common/types"
	"server/common/utils"
)

// StakersRes stakers paging return parameters
type StakersRes struct {
	Total      int64          `json:"total"`        //Total number of stakers
	Last0Total int64          `json:"last_0_total"` //The number of newly stakers in the latest 0 days (today), calculated in real time
	Last1Total int64          `json:"last_1_total"` //Number of newly stakers in the latest 1 day (yesterday), cached
	Last7Total int64          `json:"last_7_total"` //Number of newly stakers in the last 7 days, cached
	Stakers    []model.Staker `json:"stakers"`      //List of stakers
}

func FetchStakers(name string, page, size int) (res StakersRes, err error) {
	if name != "" {
		err = DB.Where("name=?", name).Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Stakers).Error
		if err != nil {
			return
		}
		err = DB.Where("name=?", name).Model(&model.Staker{}).Count(&res.Total).Error
	} else {
		err = DB.Order("block_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.Stakers).Error
		res.Total = stats.TotalStaker
	}
	if err == nil {
		res.Last0Total, err = TodayStakerTotal()
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

// TodayStakerTotal number of new stakers opened today
func TodayStakerTotal() (count int64, err error) {
	now := time.Now().Local()
	loc, _ := time.LoadLocation("Local")
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	start := startTime.Unix()
	err = DB.Model(&model.Staker{}).Where("timestamp>=? AND timestamp<?", start, now.Unix()).Count(&count).Error
	return
}

var lastStakerTime string
var last1Total int64
var last7Total int64

func cacheLastTotal() (total1 int64, total2 int64, err error) {
	if now := time.Now().Local().Format("2006-01-02"); now != lastStakerTime {
		last1Total, err = lastStakerTotal(1)
		if err != nil {
			return
		}
		last7Total, err = lastStakerTotal(7)
		if err != nil {
			return
		}
		lastStakerTime = now
	}
	return last1Total, last7Total, nil
}

// lastStakerTotal the latest number of newly stakers
func lastStakerTotal(day int64) (count int64, err error) {
	start, stop := utils.LastTimeRange(day)
	err = DB.Model(&model.Staker{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&count).Error
	return
}

func FindStaker(addr types.Address) (res model.Staker, err error) {
	err = DB.Model(&model.Staker{}).Where("address=?", addr).Scan(&res).Error
	return
}

func Stakers(page, size int, order string) (res []model.Staker, err error) {
	db := DB.Model(model.Staker{})
	if order != "" {
		db = db.Order(order)
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res).Error
	return
}

type StakerTxCountRes struct {
	Total uint64 `json:"total"`
	Day30 uint64 `json:"day30"`
	Day7  uint64 `json:"day_7"`
	Day1  uint64 `json:"day_1"`
}

func StakerTxCount(addr types.Address) (res StakerTxCountRes, err error) {
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
