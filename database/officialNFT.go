package database

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	"math/big"
	"strings"
)

// SNFT SNFT属性信息
type SNFT struct {
	Address      string  `json:"address" gorm:"type:CHAR(44);primary_key"` //SNFT地址
	CreateAt     uint64  `json:"create_at"`                                //创建时间戳
	CreateNumber uint64  `json:"create_number" gorm:"index"`               //创建的区块高度
	Creator      string  `json:"creator" gorm:"type:CHAR(44)"`             //创建者地址
	Awardee      *string `json:"awardee"`                                  //被奖励的矿工获地址
	RewardAt     *uint64 `json:"reward_at"`                                //奖励时间戳,矿工被奖励这个SNFT的时间
	RewardNumber *uint64 `json:"reward_number"`                            //奖励区块高度,矿工被奖励这个SNFT的区块高度
	Owner        *string `json:"owner" gorm:"type:CHAR(44)"`               //所有者,未分配和回收的为null
	RoyaltyRatio uint32  `json:"royalty_ratio"`                            //版税费率,单位万分之一
	MetaUrl      string  `json:"meta_url"`                                 //元信息链接
}

// RecycleSNFT SNFT回收,清空所有者
func RecycleSNFT(addr string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return recycleSNFT(tx, addr)
	})
}

// InjectSNFT 官方批量注入SNFT
func InjectSNFT(startIndex *big.Int, count uint64, royalty uint32, dir, creator string, number, timestamp uint64) (snfts, urls []string, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		SNFTAddr := big.NewInt(0)
		SNFTAddr.SetString("8000000000000000000000000000000000000000", 16)
		SNFTAddr = SNFTAddr.Add(SNFTAddr, startIndex)
		for i := uint64(0); i < count; i++ {
			addr := common.BigToAddress(SNFTAddr).String()
			// 取地址倒数3-4位作为文件名
			metaUrl := dir + addr[39:40]
			snfts = append(snfts, addr)
			urls = append(urls, metaUrl)
			err := injectSNFT(tx, strings.ToLower(addr), royalty, metaUrl, creator, number, timestamp)
			if err != nil {
				return err
			}
			SNFTAddr = SNFTAddr.Add(SNFTAddr, common.Big1)
		}
		return nil
	})
	return
}

// ImportSNFT 单个导入SNFT，主要用于创世预设的SNFT
func ImportSNFT(addr string, royalty uint32, metaUrl, creator string, number, timestamp uint64) error {
	return injectSNFT(DB, addr, royalty, metaUrl, creator, number, timestamp)
}

// SaveSNFT 更新SNFT信息
func SaveSNFT(addr string, royalty uint32, metaUrl, creator string, number, timestamp uint64) error {
	return DB.Select("address", "create_at", "create_number", "creator", "royalty_ratio", "meta_url").Save(&SNFT{
		Address:      addr,
		CreateAt:     timestamp,
		CreateNumber: number,
		Creator:      creator,
		RoyaltyRatio: royalty,
		MetaUrl:      metaUrl,
	}).Error
}

func injectSNFT(tx *gorm.DB, addr string, royalty uint32, metaUrl, creator string, number, timestamp uint64) error {
	return DB.Create(&SNFT{
		Address:      addr,
		CreateAt:     timestamp,
		CreateNumber: number,
		Creator:      creator,
		RoyaltyRatio: royalty,
		MetaUrl:      metaUrl,
	}).Error
}

func DispatchSNFT(validator, snft string, number, timestamp uint64) error {
	return DB.Select("owner", "awardee", "reward_number", "reward_at").Save(&SNFT{
		Address:      snft,
		Owner:        &validator,
		Awardee:      &validator,
		RewardNumber: &number,
		RewardAt:     &timestamp,
	}).Error
}

func FetchSNFTs(owner string, page, size uint64) (data []SNFT, count int64, err error) {
	if owner != "" {
		err = DB.Where("owner=?", owner).Order("create_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Where("owner=?", owner).Model(&SNFT{}).Count(&count).Error
	} else {
		err = DB.Order("create_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&SNFT{}).Count(&count).Error
	}
	return
}

func BlockSNFTs(number uint64) (data []SNFT, err error) {
	err = DB.Where("reward_number=?", number).Find(&data).Error
	return
}

func recycleSNFT(tx *gorm.DB, addr string) error {
	if len(addr) == 42 {
		return tx.Model(&SNFT{}).Where("address=?", addr).UpdateColumns(map[string]interface{}{
			"owner":         nil,
			"Awardee":       nil,
			"reward_at":     nil,
			"reward_number": nil,
		}).Error
	} else {
		for i := 0; i < 16; i++ {
			err := recycleSNFT(tx, fmt.Sprintf("%s%x", addr, i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
