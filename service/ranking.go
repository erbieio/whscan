package service

import (
	"server/common/model"
)

// RankingSNFTRes SNFT ranking return parameters
type RankingSNFTRes struct {
	Total int64        `json:"total"` //The total number of SNFTs
	Data  []model.SNFT `json:"data"`  //SNFT list
}

func RankingSNFT(page, size int) (res RankingSNFTRes, err error) {
	db := DB.Model(&model.SNFT{}).Order("tx_amount DESC").Where("tx_amount!=0")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}

// RankingNFTRes NFT ranking return parameters
type RankingNFTRes struct {
	Total int64       `json:"total"` //The total number of NFTs
	Data  []model.NFT `json:"data"`  //NFT list
}

func RankingNFT(page, size int) (res RankingNFTRes, err error) {
	db := DB.Model(&model.NFT{}).Order("tx_amount DESC").Where("tx_amount!=0")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	return
}

// RankingStakerRes Staker ranking return parameters
type RankingStakerRes struct {
	Total int64          `json:"total"` //The total number of Staker
	Data  []model.Staker `json:"data"`  //Staker list
}

func RankingStaker(page, size int) (res RankingStakerRes, err error) {
	db := DB.Model(&model.Staker{}).Order("reward DESC").Where("reward!=0")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Scan(&res.Data).Error
	return
}
