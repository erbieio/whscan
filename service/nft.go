package service

import "server/common/model"

// UNFTsRes NFT paging return parameters
type UNFTsRes struct {
	Total int64        `json:"total"` //The total number of NFTs
	NFTs  []model.UNFT `json:"nfts"`  //NFT list
}

func FetchUNFTs(exchanger, collectionId, owner string, page, size int) (res UNFTsRes, err error) {
	db := DB.Model(&model.UNFT{})
	if exchanger != "" {
		db = db.Where("exchanger_addr=?", exchanger)
	}
	if collectionId != "" {
		db = db.Where("collection_id=?", collectionId)
	}
	if owner != "" {
		db = db.Where("owner=?", owner)
	}

	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

// NFTTxsRes NFT transaction paging return parameters
type NFTTxsRes struct {
	Total  int64         `json:"total"`   //The total number of NFTs
	NFTTxs []model.NFTTx `json:"nft_txs"` //NFT transaction list
}

func FetchNFTTxs(address, exchanger, account string, page, size int) (res NFTTxsRes, err error) {
	db := DB
	if exchanger != "" {
		db = db.Where("exchanger_addr=?", exchanger)
	}
	if account != "" {
		db = db.Where("`from`=? OR `to`=?", account, account)
	}
	if address != "" {
		db = db.Where("nft_addr=?", address)
	}

	err = db.Model(&model.NFTTx{}).Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Order("timestamp DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTTxs).Error
	return
}

// SNFTsRes SNFT paging return parameters
type SNFTsRes struct {
	Total int64        `json:"total"` //The total number of SNFTs
	NFTs  []model.SNFT `json:"nfts"`  //SNFT list
}

func FetchSNFTs(owner string, page, size int) (res SNFTsRes, err error) {
	db := DB
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if owner == "" {
		res.Total = int64(TotalSNFT())
	} else {
		err = db.Model(&model.SNFT{}).Count(&res.Total).Error
	}
	if err != nil {
		return
	}
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
	return
}

// SNFTsAndMetaRes SNFT and meta information paging return parameters
type SNFTsAndMetaRes struct {
	Total int64 `json:"total"` //The total number of SNFTs
	NFTs  []struct {
		model.SNFT
		model.FullNFT
	} `json:"nfts"` //SNFT list
}

func FetchSNFTsAndMeta(owner, collectionId string, page, size int) (res SNFTsAndMetaRes, err error) {
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN full_nfts ON LEFT(address,40)=full_nfts.id")
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if collectionId != "" {
		db = db.Where("group_id=?", collectionId)
	}
	if owner == "" && collectionId == "" {
		res.Total = int64(TotalSNFT())
	} else {
		err = db.Count(&res.Total).Error
	}
	if err != nil {
		return
	}
	err = db.Select("snfts.*,full_nfts.*").Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

func BlockSNFTs(number uint64) (res []model.SNFT, err error) {
	err = DB.Where("reward_number=?", number).Find(&res).Error
	return
}

// SNFTGroupsRes SNFT collection information return
type SNFTGroupsRes struct {
	Total       int64 `json:"total"` //The total number of SNFT collections
	Collections []struct {
		model.Collection
		TotalHold int64 `json:"total_hold"` //The number of SNFTs held in a collection
		FullNFTs  []struct {
			model.FullNFT
			TotalHold int64 `json:"total_hold"` //The number of SNFTs held in a FullNFT
		} `gorm:"-"` // 16 FullNFT messages
	} `json:"collections"` //collection information
}

func FindSNFTGroups(owner string, page, size int) (res SNFTGroupsRes, err error) {
	err = DB.Model(&model.SNFT{}).Where("owner=?", owner).Select("COUNT(DISTINCT LEFT(address, 39))").Scan(&res.Total).Error
	if err != nil {
		return
	}
	err = DB.Model(&model.SNFT{}).
		Joins("LEFT JOIN collections on LEFT(address,39) = id").
		Select("`collections`.*,COUNT(address) AS total_hold").Where("owner=?", owner).Group("id").
		Order("id DESC").Offset((page - 1) * size).Limit(size).Scan(&res.Collections).Error
	for i := range res.Collections {
		err = DB.Model(&model.SNFT{}).
			Joins("LEFT JOIN full_nfts on LEFT(address,40) = full_nfts.id").
			Select("full_nfts.*,COUNT(*) AS total_hold").
			Where("LEFT(address, 39)=? AND owner=?", res.Collections[i].Id, owner).Group("id").
			Scan(&res.Collections[i].FullNFTs).Error
		if err != nil {
			return
		}
	}
	return
}

func FullSNFTs(fullNFTId string) (res []model.SNFT, err error) {
	err = DB.Where("LEFT(address, 40)=?", fullNFTId).Find(&res).Error
	return
}
