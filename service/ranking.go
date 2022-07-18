package service

import (
	"server/common/model"
	"server/common/utils"
)

// RankingSNFTRes SNFT ranking return parameters
type RankingSNFTRes struct {
	Total uint64 `json:"total"` //The total number of SNFTs
	NFTs  []struct {
		model.SNFT
		Creator string `json:"creator"`
		TxCount uint64 `json:"txCount"`
	} `json:"nfts"` //SNFT list
}

func RankingSNFT(limit string, page, size int) (res RankingSNFTRes, err error) {
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN epoches ON LEFT(address,38)=epoches.id").
		Group("address,creator").Offset((page - 1) * size).Limit(size).
		Order("tx_count DESC, address DESC").Select("snfts.*,creator,COUNT(nft_addr) AS tx_count")
	switch limit {
	case "24h":
		start, stop := utils.LastTimeRange(1)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "7d":
		start, stop := utils.LastTimeRange(7)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "30d":
		start, stop := utils.LastTimeRange(30)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	default:
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address")
	}
	err = db.Scan(&res.NFTs).Error
	res.Total = cache.TotalSNFT
	return
}

// RankingNFTRes NFT ranking return parameters
type RankingNFTRes struct {
	Total uint64 `json:"total"` //The total number of NFTs
	NFTs  []struct {
		model.NFT
		TxCount uint64 `json:"txCount"`
	} `json:"nfts"` //NFT list
}

func RankingNFT(limit string, page, size int) (res RankingNFTRes, err error) {
	db := DB.Model(&model.NFT{}).Group("address").Offset((page - 1) * size).Limit(size).
		Order("tx_count DESC, address DESC").Select("nfts.*,COUNT(nft_addr) AS tx_count")
	switch limit {
	case "24h":
		start, stop := utils.LastTimeRange(1)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "7d":
		start, stop := utils.LastTimeRange(7)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	case "30d":
		start, stop := utils.LastTimeRange(30)
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address AND nft_txes.timestamp>=? AND nft_txes.timestamp<?", start, stop)
	default:
		db = db.Joins("LEFT JOIN nft_txes ON nft_addr=address")
	}
	err = db.Scan(&res.NFTs).Error
	res.Total = cache.TotalNFT
	return
}

// RankingExchangerRes Exchanger ranking return parameters
type RankingExchangerRes struct {
	Total      uint64 `json:"total"` //The total number of Exchanger
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
	res.Total = cache.TotalExchanger
	return
}
