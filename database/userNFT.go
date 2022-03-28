package database

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// UserNFT 用户NFT属性信息
type UserNFT struct {
	Address       string `json:"address" gorm:"type:CHAR(44);primary_key"` //NFT地址
	RoyaltyRatio  uint32 `json:"royalty_ratio"`                            //版税费率,单位万分之一
	MetaUrl       string `json:"meta_url"`                                 //元信息URL
	ExchangerAddr string `json:"exchanger_addr" gorm:"type:CHAR(44)"`      //所在交易所地址,没有的可以在任意交易所交易
	Creator       string `json:"creator" gorm:"type:CHAR(44)"`             //创建者地址
	Timestamp     uint64 `json:"timestamp"`                                //创建时间戳
	BlockNumber   uint64 `json:"block_number"`                             //创建的区块高度
	TxHash        string `json:"tx_hash" gorm:"type:CHAR(66)"`             //创建的交易哈希
	Owner         string `json:"owner" gorm:"type:CHAR(44)"`               //所有者
}

// userNFTCount 值为NFT总数加一
var userNFTCount *big.Int

// Insert 更新NFT地址（加一）再插入NFT信息
func (un UserNFT) Insert() error {
	if err := updateNFTCount(); err != nil {
		return err
	}
	return DB.Create(&un).Error
}

func FetchUserNFTs(page, size int) (data []UserNFT, count int, err error) {
	err = DB.Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
	count = len(data)
	return
}

func FindUserNFT(addr string) (data UserNFT, err error) {
	err = DB.Where("address=?", addr).First(&data).Error
	return
}

// initNFTCount 初始化NFT数量，值为NFT总数加一
func initNFTCount() (err error) {
	if userNFTCount == nil {
		var count int64
		err = DB.Model(&UserNFT{}).Count(&count).Error
		userNFTCount = big.NewInt(count + 1)
	}
	return
}

// 更新NFT数量（加一）
func updateNFTCount() error {
	if err := initNFTCount(); err != nil {
		return err
	}
	userNFTCount = userNFTCount.Add(userNFTCount, common.Big1)
	return nil
}

// GetNFTAddr 获取NFT地址
func GetNFTAddr() (string, error) {
	if err := initNFTCount(); err != nil {
		return "", err
	}
	return common.BigToAddress(userNFTCount).String(), nil
}
