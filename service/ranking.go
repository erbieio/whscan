package service

import (
	"server/common/model"
	"server/common/utils"
)

// RankingSNFTRes SNFT ranking return parameters
type RankingSNFTRes struct {
	Total int64 `json:"total"` //The total number of SNFTs
	NFTs  []*struct {
		model.SNFT
		model.FNFT
		Creator string `json:"creator"`
		TxCount uint64 `json:"txCount"`
	} `json:"nfts"` //SNFT list
}

func RankingSNFT(limit string, page, size int) (res RankingSNFTRes, err error) {
	db := DB.Model(&model.NFTTx{}).Joins("LEFT JOIN epoches ON LEFT(nft_addr,39)=epoches.id").
		Joins("LEFT JOIN fnfts ON LEFT(nft_addr,41)=fnfts.id").
		Joins("LEFT JOIN snfts ON nft_addr=address").
		Order("tx_count DESC, nft_addr DESC").Select("snfts.*,fnfts.*,creator,COUNT(nft_addr) AS tx_count")
	switch limit {
	case "24h":
		start, stop := utils.LastTimeRange(1)
		db = db.Where("LEFT(address,3)='0x8' AND epoches.timestamp>=? AND epoches.timestamp<?", start, stop)
	case "7d":
		start, stop := utils.LastTimeRange(7)
		db = db.Where("LEFT(address,3)='0x8' AND epoches.timestamp>=? AND epoches.timestamp<?", start, stop)
	case "30d":
		start, stop := utils.LastTimeRange(30)
		db = db.Where("LEFT(address,3)='0x8' AND epoches.timestamp>=? AND epoches.timestamp<?", start, stop)
	default:
		db = db.Where("LEFT(address,3)='0x8'")
	}
	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Group("nft_addr,epoches.id,fnfts.id,address").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

// RankingNFTRes NFT ranking return parameters
type RankingNFTRes struct {
	Total int64 `json:"total"` //The total number of NFTs
	NFTs  []struct {
		model.NFT
		TxCount uint64 `json:"txCount"`
	} `json:"nfts"` //NFT list
}

func RankingNFT(limit string, page, size int) (res RankingNFTRes, err error) {
	db := DB.Model(&model.NFTTx{}).Joins("LEFT JOIN nfts ON nft_addr=address").Order("tx_count DESC, address DESC").Select("nfts.*,COUNT(nft_addr) AS tx_count")
	switch limit {
	case "24h":
		start, stop := utils.LastTimeRange(1)
		db = db.Where("LEFT(address,3)='0x0' AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "7d":
		start, stop := utils.LastTimeRange(7)
		db = db.Where("LEFT(address,3)='0x0' AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "30d":
		start, stop := utils.LastTimeRange(30)
		db = db.Where("LEFT(address,3)='0x0' AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	default:
		db = db.Where("LEFT(address,3)='0x0'")
	}
	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Group("address").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

// RankingExchangerRes Exchanger ranking return parameters
type RankingExchangerRes struct {
	Total      int64 `json:"total"` //The total number of Exchanger
	Exchangers []struct {
		model.Exchanger
		TxCount uint64 `json:"txCount"`
	} `json:"exchangers"` //Exchanger list
}

func RankingExchanger(limit string, page, size int) (res RankingExchangerRes, err error) {
	db := DB.Model(&model.Exchanger{}).Group("address").Offset((page - 1) * size).Limit(size).
		Order("tx_count DESC, address DESC").Select("exchangers.*,COUNT(nft_addr) AS tx_count")
	switch limit {
	case "24h":
		start, stop := utils.LastTimeRange(1)
		db = db.Joins("LEFT JOIN nft_txes ON exchanger_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "7d":
		start, stop := utils.LastTimeRange(7)
		db = db.Joins("LEFT JOIN nft_txes ON exchanger_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "30d":
		start, stop := utils.LastTimeRange(30)
		db = db.Joins("LEFT JOIN nft_txes ON exchanger_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	default:
		db = db.Joins("LEFT JOIN nft_txes ON exchanger_addr=address")
	}
	err = db.Scan(&res.Exchangers).Error
	res.Total = stats.TotalExchanger
	return
}
