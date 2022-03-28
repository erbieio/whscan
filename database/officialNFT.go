package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"math/big"
)

// OfficialNFT SNFT属性信息
type OfficialNFT struct {
	Address     string `json:"address" gorm:"type:CHAR(44);primary_key"` //SNFT地址
	Timestamp   uint64 `json:"timestamp"`                                //创建时间戳
	BlockNumber uint64 `json:"block_number" gorm:"index"`                //创建的区块高度
	Owner       string `json:"owner" gorm:"type:CHAR(44)"`               //所有者
}

// SNFTRecycler SNFT兑换回收
type SNFTRecycler struct {
	ID      uint64 `gorm:"primary_key;auto_increment"` //编号
	Address string ` gorm:"type:CHAR(44)"`             //SNFT地址
}

//maxSNFTAddr 可用SNFT地址（新的地址是之前地址加一）
var maxSNFTAddr *big.Int
var sNFTPools []string

// SNFTRecycle SNFT回收,删除NFT，将地址放回池子
func SNFTRecycle(addr string) error {
	if err := initSNFTPools(); err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		return sNFTRecycle(addr)
	})
}

func sNFTRecycle(addr string) error {
	if len(addr) == 42 {
		// 加入池子并写到数据库
		sNFTPools = append(sNFTPools, addr)
		return DB.Create(&SNFTRecycler{Address: addr}).Error
	} else {
		for i := 0; i < 16; i++ {
			err := SNFTRecycle(fmt.Sprintf("%s%x", addr, i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// initNFTAddr 初始化NFT地址，值为NFT总数加一
func initSNFTPools() (err error) {
	if sNFTPools == nil || maxSNFTAddr == nil {
		err = DB.Model(&SNFTRecycler{}).Order("id ASC").Select("address").Find(&sNFTPools).Error
		if err != nil {
			return
		}
		err = DB.Select("MAX(address)").Find(&maxSNFTAddr).Error
	}
	return
}
