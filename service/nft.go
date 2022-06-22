package service

import "server/common/model"

// UserNFTsRes NFT分页返回参数
type UserNFTsRes struct {
	Total int64           `json:"total"` //NFT总数
	NFTs  []model.UserNFT `json:"nfts"`  //NFT列表
}

func FetchUserNFTs(exchanger, collectionId, owner string, page, size int) (res UserNFTsRes, err error) {
	db := DB.Model(&model.UserNFT{})
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
	Total int64        `json:"total"` //SNFT总数
	NFTs  []model.SNFT `json:"nfts"`  //SNFT列表
}

func FetchSNFTs(owner string, page, size int) (res SNFTsRes, err error) {
	db := DB
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if owner == "" {
		res.Total = int64(TotalOfficialNFT())
	} else {
		err = db.Model(&model.SNFT{}).Count(&res.Total).Error
	}
	if err != nil {
		return
	}
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
	return
}

// SNFTsAndMetaRes SNFT和元信息分页返回参数
type SNFTsAndMetaRes struct {
	Total int64 `json:"total"` //SNFT总数
	NFTs  []struct {
		model.SNFT
		model.FullNFT
	} `json:"nfts"` //SNFT列表
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
		res.Total = int64(TotalOfficialNFT())
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

// SNFTGroupsRes SNFT合集信息返回
type SNFTGroupsRes struct {
	Total       int64 `json:"total"` //SNFT合集总数
	Collections []struct {
		model.Collection
		TotalHold int64 `json:"total_hold"` //一个合集里的持有SNFT数量
		FullNFTs  []struct {
			model.FullNFT
			TotalHold int64 `json:"total_hold"` //一个FullNFT里的持有SNFT数量
		} `gorm:"-"` // 16个FullNFT信息
	} `json:"collections"` //合集信息
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
