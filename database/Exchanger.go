package database

import "github.com/jinzhu/gorm"

// Exchanger 交易所属性信息
type Exchanger struct {
	Address     string `json:"address" gorm:"type:CHAR(42);primaryKey"` //交易所地址
	Name        string `json:"name"`                                    //交易所名称
	URL         string `json:"url"`                                     //交易所URL
	FeeRatio    uint32 `json:"fee_ratio"`                               //手续费率,单位万分之一
	Creator     string `json:"creator" gorm:"type:CHAR(42)"`            //创建者地址
	Timestamp   uint64 `json:"timestamp"`                               //开启时间
	IsOpen      bool   `json:"is_open"`                                 //是否开启中
	BlockNumber uint64 `json:"block_number" gorm:"index:,sort:DESC"`    //创建时的区块号
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66)"`            //创建的交易
}

func OpenExchange(e *Exchanger) error {
	oldE := &Exchanger{}
	err := DB.Where("address=?", e.Address).First(oldE).Error
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

func FetchExchangers(page, size int) (data []Exchanger, count int, err error) {
	err = DB.Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
	count = len(data)
	return
}

func FindExchanger(addr string) (data Exchanger, err error) {
	err = DB.Where("address=?", addr).First(&data).Error
	return
}
