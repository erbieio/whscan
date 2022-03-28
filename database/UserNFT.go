package database

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// UserNFT 用户NFT属性信息
type UserNFT struct {
	Address       string `json:"address" gorm:"type:CHAR(44);primaryKey"` //NFT地址
	RoyaltyRatio  uint32 `json:"royalty_ratio"`                           //版税费率,单位万分之一
	MetaUrl       string `json:"meta_url"`                                //元信息URL
	ExchangerAddr string `json:"exchanger_addr" gorm:"type:CHAR(44)"`     //所在交易所地址,没有的可以在任意交易所交易
	Creator       string `json:"creator" gorm:"type:CHAR(44)"`            //创建者地址
	Timestamp     uint64 `json:"timestamp"`                               //创建时间戳
	BlockNumber   uint64 `json:"block_number"`                            //创建的区块高度
	TxHash        string `json:"tx_hash" gorm:"type:CHAR(66)"`            //创建的交易哈希
	Owner         string `json:"owner" gorm:"type:CHAR(44)"`              //所有者
}

// nftAddr 可用NFT地址（新的NFT地址是之前NFT的总数加一）
var nftAddr *big.Int

// Insert 更新NFT地址（加一）再插入NFT信息
func (un UserNFT) Insert() error {
	if err := updateNFTAddr(); err != nil {
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

// initNFTAddr 初始化NFT地址，值为NFT总数加一
func initNFTAddr() (err error) {
	if nftAddr == nil {
		var count int64
		err = DB.Model(&UserNFT{}).Count(&count).Error
		nftAddr = big.NewInt(count + 1)
	}
	return
}

// 更新NFT地址（加一）
func updateNFTAddr() error {
	if err := initNFTAddr(); err != nil {
		return err
	}
	nftAddr = nftAddr.Add(nftAddr, common.Big1)
	return nil
}

// GetNFTAddr 获取NFT地址
func GetNFTAddr() (string, error) {
	if err := initNFTAddr(); err != nil {
		return "", err
	}
	return common.BigToAddress(nftAddr).String(), nil
}
