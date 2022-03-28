package database

import (
	"errors"
	"math/big"
)

// Pledge 帐户质押金额
type Pledge struct {
	Address string `json:"address" gorm:"type:CHAR(44);primary_key"` //质押帐户
	Amount  string `json:"amount" gorm:"type:CHAR(64)"`              //质押金额
	Count   uint64 `json:"count"`                                    //质押次数，PledgeAdd和PledgeSub都加一次
}

type ExchangerPledge Pledge //交易所质押
type ConsensusPledge Pledge //共识质押

// ExchangerPledgeAdd 增加质押金额
func ExchangerPledgeAdd(addr, amount string) error {
	pledge := ExchangerPledge{}
	err := DB.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		pledge.Amount = "0x0"
	}
	pledge.Amount = AddAndPadding(pledge.Amount, amount)
	return DB.Create(&pledge).Error
}

// ExchangerPledgeSub 减少质押金额
func ExchangerPledgeSub(addr, amount string) error {
	pledge := ExchangerPledge{}
	err := DB.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		return errors.New(addr + "没有质押过")
	}
	pledge.Amount = AddAndPadding(pledge.Amount, "-"+amount)
	return DB.Create(&pledge).Error
}

// ConsensusPledgeAdd 增加质押金额
func ConsensusPledgeAdd(addr, amount string) error {
	pledge := ConsensusPledge{}
	err := DB.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		pledge.Amount = "0x0"
	}
	pledge.Amount = AddAndPadding(pledge.Amount, amount)
	return DB.Create(&pledge).Error
}

// ConsensusPledgeSub 减少质押金额
func ConsensusPledgeSub(addr, amount string) error {
	pledge := ConsensusPledge{}
	err := DB.Where("address=?", addr).Find(&pledge).Error
	if err != nil {
		return err
	}
	pledge.Count++
	pledge.Address = addr
	if pledge.Amount == "" {
		return errors.New(addr + "没有质押过")
	}
	pledge.Amount = AddAndPadding(pledge.Amount, "-"+amount)
	return DB.Create(&pledge).Error
}

var padding = "00000000000000000000000000000000000000000000000000000000000000"

// AddAndPadding 两个16进制大数字符串相加，并左0到64位
func AddAndPadding(a, b string) string {
	aa, ok := new(big.Int).SetString(a, 16)
	if !ok {
		panic(nil)
	}
	bb, ok := new(big.Int).SetString(b, 16)
	if !ok {
		panic(nil)
	}
	cc := aa.Add(aa, bb)
	if cc.Sign() == -1 {
		panic(nil)
	}
	ccStr := cc.Text(16)
	return "0x" + padding[0:64-len(ccStr)] + ccStr[2:len(ccStr)-2]
}
