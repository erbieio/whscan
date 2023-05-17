package service

import (
	"server/common/model"
)

// RankingSNFTRes SNFT ranking return parameters
type RankingSNFTRes struct {
	Total int64 `json:"total"` //The total number of SNFTs
	NFTs  []*struct {
		model.SNFT
		model.FNFT
	} `json:"nfts"` //SNFT list
}

func RankingSNFT(page, size int) (res RankingSNFTRes, err error) {
	db := DB.Model(&model.SNFT{}).Order("tx_amount DESC").Where("tx_amount!='0'")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Joins("LEFT JOIN fnfts ON LEFT(address,41)=fnfts.id").Offset((page - 1) * size).Limit(size).Select("*").Scan(&res.NFTs).Error
	return
}

// RankingNFTRes NFT ranking return parameters
type RankingNFTRes struct {
	Total int64       `json:"total"` //The total number of NFTs
	NFTs  []model.NFT `json:"nfts"`  //NFT list
}

func RankingNFT(page, size int) (res RankingNFTRes, err error) {
	db := DB.Model(&model.NFT{}).Order("tx_amount DESC").Where("tx_amount!='0'")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
	return
}

// RankingExchangerRes Exchanger ranking return parameters
type RankingExchangerRes struct {
	Total      int64              `json:"total"`      //The total number of Exchanger
	Exchangers []*model.Exchanger `json:"exchangers"` //Exchanger list
}

func RankingExchanger(page, size int) (res RankingExchangerRes, err error) {
	db := DB.Model(&model.Exchanger{}).Order("tx_amount DESC").Where("tx_amount!='0'")
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Exchangers).Error
	return
}
