package database

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Exchanger 交易所属性信息
type Exchanger struct {
	Address     string `json:"address" gorm:"type:CHAR(42);primary_key"` //交易所地址
	Name        string `json:"name" gorm:"type:VARCHAR(256)"`            //交易所名称
	URL         string `json:"url"`                                      //交易所URL
	FeeRatio    uint32 `json:"fee_ratio"`                                //手续费率,单位万分之一
	Creator     string `json:"creator" gorm:"type:CHAR(42)"`             //创建者地址
	Timestamp   uint64 `json:"timestamp" gorm:"index"`                   //开启时间
	IsOpen      bool   `json:"is_open"`                                  //是否开启中
	BlockNumber uint64 `json:"block_number" gorm:"index"`                //创建时的区块号
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66)"`             //创建的交易
	NFTCount    uint64 `json:"nft_count" gorm:"-"`                       //NFT总数，批量查询的此字段无效
}

func OpenExchange(e *Exchanger) error {
	err := DB.Where("address=?", e.Address).First(&Exchanger{}).Error
	if gorm.IsRecordNotFoundError(err) {
		return DB.Create(e).Error
	}
	return DB.Model(&Exchanger{}).Where("address=?", e.Address).Updates(map[string]interface{}{
		"is_open":   true,
		"name":      e.Name,
		"url":       e.URL,
		"fee_ratio": e.FeeRatio,
	}).Error
}

func CloseExchange(addr string) error {
	return DB.Model(&Exchanger{}).Where("address=?", addr).Update("is_open", false).Error
}

func FetchExchangers(name string, page, size uint64) (data []Exchanger, count int64, err error) {
	if name != "" {
		err = DB.Where("name=?", name).Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Where("name=?", name).Model(&Exchanger{}).Count(&count).Error
	} else {
		err = DB.Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&Exchanger{}).Count(&count).Error
	}
	return
}

// YesterdayExchangerTotal 昨日创建的交易所数量
func YesterdayExchangerTotal() (count int64, err error) {
	now := time.Now().Local()
	loc, _ := time.LoadLocation("Local")
	daySecond := 24 * time.Hour.Milliseconds() / 1000
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02")+" 00:00:00", loc)
	stop := startTime.Unix()
	start := stop - daySecond
	err = DB.Model(&Exchanger{}).Where("timestamp>=? AND timestamp<?", start, stop).Count(&count).Error
	return
}

func FindExchanger(addr string) (data Exchanger, err error) {
	err = DB.Where("address=?", addr).First(&data).Error
	if err != nil {
		return
	}
	err = DB.Where("exchanger_addr=?", addr).Model(&UserNFT{}).Count(&data.NFTCount).Error
	return
}
