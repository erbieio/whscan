package database

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
)

// OfficialNFT SNFT属性信息
type OfficialNFT struct {
	Address     string `json:"address" gorm:"type:CHAR(44);primary_key"` //SNFT地址
	Timestamp   uint64 `json:"timestamp"`                                //创建时间戳
	BlockNumber uint64 `json:"block_number" gorm:"index"`                //创建的区块高度
	Owner       string `json:"owner" gorm:"type:CHAR(44)"`               //所有者
}

// RecycleSNFT SNFT兑换回收
type RecycleSNFT struct {
	ID      uint64 `gorm:"primary_key;auto_increment"` //编号
	Address string ` gorm:"type:CHAR(44)"`             //SNFT地址
}

// EnqueueSNFT SNFT入队回收，将地址放到回收池子
func EnqueueSNFT(addr string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return recycleSNFT(addr)
	})
}

func DispatchSNFT(validators []string, blockNumber, timestamp uint64) error {
	for _, validator := range validators {
		SNFT, err := dequeueSNFT()
		if err != nil {
			return err
		}
		DB.Create(OfficialNFT{
			Address:     SNFT,
			Timestamp:   timestamp,
			BlockNumber: blockNumber,
			Owner:       validator,
		})
	}
	return nil
}

func dequeueSNFT() (addr string, err error) {
	if err = initSNFTCount(); err != nil {
		return "", err
	}
	// 否则返回可用SNFT地址，并将可用SNFT地址加一
	var count int64
	err = DB.Model(&RecycleSNFT{}).Count(&count).Error
	if err != nil {
		return
	}
	if count > 0 {
		// 从回收池里取出id最小的一个并删除
		recycleSNFT := RecycleSNFT{}
		err = DB.Order("id ASC").Limit(1).Find(&recycleSNFT).Error
		if err != nil {
			return
		}
		addr = recycleSNFT.Address
		err = DB.Delete(&recycleSNFT).Error
	} else {
		addr = common.BigToAddress(SNFTCount).String()
		SNFTCount = SNFTCount.Add(SNFTCount, common.Big1)
	}
	return
}

func recycleSNFT(addr string) error {
	if len(addr) == 42 {
		// 删除SNFT并将地址写到SNFT回收表里
		err := DB.Delete(&OfficialNFT{}, "address=?", addr).Error
		if err != nil {
			return err
		}
		return DB.Create(&RecycleSNFT{Address: addr}).Error
	} else {
		for i := 0; i < 16; i++ {
			err := recycleSNFT(fmt.Sprintf("%s%x", addr, i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//SNFTCount SNFT数量（包含回收池里的）
var SNFTCount *big.Int

// initSNFTCount 初始化SNFT数量
func initSNFTCount() (err error) {
	if SNFTCount == nil {
		var count1, count2 int64
		err = DB.Model(&OfficialNFT{}).Count(&count1).Error
		if err != nil {
			return
		}
		err = DB.Model(&RecycleSNFT{}).Count(&count2).Error
		if err != nil {
			return
		}
		SNFTCount = big.NewInt(0)
		SNFTCount.SetString("8000000000000000000000000000000000000000", 16)
		SNFTCount = SNFTCount.Add(SNFTCount, big.NewInt(count1+count2))
	}
	return
}
