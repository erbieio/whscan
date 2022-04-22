package model

// Pledge 账户质押金额
type Pledge struct {
	Address string `json:"address" gorm:"type:CHAR(44);primary_key"` //质押账户
	Amount  string `json:"amount" gorm:"type:CHAR(64)"`              //质押金额
	Count   uint64 `json:"count"`                                    //质押次数，PledgeAdd和PledgeSub都加一次
}

// ExchangerPledge 交易所质押
type ExchangerPledge Pledge

// ConsensusPledge 共识质押
type ConsensusPledge Pledge
