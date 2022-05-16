package service

import "server/common/model"

// UserNFTsRes NFT分页返回参数
type UserNFTsRes struct {
	Total int64           `json:"total"` //NFT总数
	NFTs  []model.UserNFT `json:"nfts"`  //NFT列表
}

func FetchUserNFTs(exchanger, owner string, page, size int) (res UserNFTsRes, err error) {
	db := DB
	if exchanger != "" {
		db = db.Where("exchanger_addr=?", exchanger)
	}
	if owner != "" {
		db = db.Where("owner=?", owner)
	}

	err = db.Model(&model.UserNFT{}).Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
	return
}

// UserNFTsAndMetaRes NFT和元信息分页返回参数
type UserNFTsAndMetaRes struct {
	Total int64 `json:"total"` //NFT总数
	NFTs  []struct {
		model.UserNFT
		model.NFTMeta
	} `json:"nfts"` //NFT列表
}

func FetchUserNFTsAndMeta(exchanger, collectionId, owner string, page, size int) (res UserNFTsAndMetaRes, err error) {
	db := DB.Model(&model.UserNFT{}).Joins("LEFT JOIN nft_meta ON user_nfts.address=nft_meta.nft_addr")
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
	err = db.Select("user_nfts.*,nft_meta.*").Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

// NFTTxsRes NFT交易分页返回参数
type NFTTxsRes struct {
	Total  int64         `json:"total"`   //NFT总数
	NFTTxs []model.NFTTx `json:"nft_txs"` //NFT交易列表
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

// SNFTsRes SNFT分页返回参数
type SNFTsRes struct {
	Total int64               `json:"total"` //SNFT总数
	NFTs  []model.OfficialNFT `json:"nfts"`  //SNFT列表
}

func FetchSNFTs(owner string, page, size int) (res SNFTsRes, err error) {
	db := DB
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if owner == "" {
		res.Total = int64(TotalOfficialNFT())
	} else {
		err = db.Model(&model.OfficialNFT{}).Count(&res.Total).Error
	}
	if err != nil {
		return
	}
	err = db.Order("create_number DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
	return
}

// SNFTsAndMetaRes SNFT和元信息分页返回参数
type SNFTsAndMetaRes struct {
	Total int64 `json:"total"` //SNFT总数
	NFTs  []struct {
		model.OfficialNFT
		model.NFTMeta
	} `json:"nfts"` //SNFT列表
}

func FetchSNFTsAndMeta(owner, collectionId string, page, size int) (res SNFTsAndMetaRes, err error) {
	db := DB.Model(&model.OfficialNFT{}).Joins("LEFT JOIN nft_meta ON official_nfts.address=nft_meta.nft_addr")
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if collectionId != "" {
		db = db.Where("collection_id=?", collectionId)
	}
	if owner == "" && collectionId == "" {
		res.Total = int64(TotalOfficialNFT())
	} else {
		if collectionId == "" {
			err = DB.Model(&model.OfficialNFT{}).Where("owner=?", owner).Count(&res.Total).Error
		} else {
			err = db.Count(&res.Total).Error
		}
	}
	if err != nil {
		return
	}
	err = db.Select("official_nfts.*,nft_meta.*").Order("create_number DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

func BlockSNFTs(number uint64) (data []model.OfficialNFT, err error) {
	err = DB.Where("reward_number=?", number).Find(&data).Error
	return
}
