package service

import "server/common/model"

// NFTsRes NFT paging return parameters
type NFTsRes struct {
	Total int64 `json:"total"` //The total number of NFTs
	NFTs  []*struct {
		model.NFT
		CollectionName string `json:"collectionName"`
	} `json:"nfts"` //NFT list
}

func FetchNFTs(exchanger, collectionId, owner string, page, size int) (res NFTsRes, err error) {
	db := DB.Model(&model.NFT{}).Joins("LEFT JOIN collections ON id=LEFT(address,39)")
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
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Select("nfts.*, collections.name AS collection_name").Scan(&res.NFTs).Error
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
		model.FNFT
		Exchanger      string `json:"exchanger"`
		CollectionName string `json:"collectionName"`
	} `json:"nfts"` //SNFT list
}

func FetchSNFTsAndMeta(owner, exchanger, collectionId string, page, size int) (res SNFTsAndMetaRes, err error) {
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN fnfts ON LEFT(address,40)=fnfts.id").
		Joins("LEFT JOIN epoches ON LEFT(address,38)=epoches.id").
		Joins("LEFT JOIN collections ON collections.id=LEFT(address,39)")
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if collectionId != "" {
		db = db.Where("group_id=?", collectionId)
	}
	if exchanger != "" {
		db = db.Where("LEFT(address,38) IN (?)", DB.Model(&Epoch{}).Where("exchanger=?", exchanger).Select("id"))
	}
	if owner == "" && collectionId == "" && exchanger == "" {
		res.Total = int64(TotalSNFT())
	} else {
		err = db.Count(&res.Total).Error
	}
	if err != nil {
		return
	}
	err = db.Select("snfts.*,fnfts.*,epoches.exchanger,collections.name AS collection_name").Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

type NFT struct {
	model.NFT
	model.Collection
}

func GetNFT(addr string) (res NFT, err error) {
	err = DB.Model(&model.NFT{}).Joins("LEFT JOIN collections ON collection_id=collections.id").
		Where("address=?", addr).Select("nfts.*,collections.*").Scan(&res).Error
	return
}

type SNFT struct {
	model.SNFT
	model.FNFT
	model.Epoch
}

func GetSNFT(addr string) (res SNFT, err error) {
	err = DB.Model(&model.SNFT{}).Joins("LEFT JOIN fnfts ON LEFT(address,40)=fnfts.id").
		Joins("LEFT JOIN epoches ON LEFT(address,38)=epoches.id").
		Where("address=?", addr).Select("snfts.*,fnfts.*,epoches.*").Scan(&res).Error
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
			model.FNFT
			TotalHold int64 `json:"total_hold"` //The number of SNFTs held in a FNFT
		} `gorm:"-"` // 16 FNFT messages
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
			Joins("LEFT JOIN fnfts on LEFT(address,40) = fnfts.id").
			Select("fnfts.*,COUNT(*) AS total_hold").
			Where("LEFT(address, 39)=? AND owner=?", res.Collections[i].Id, owner).Group("id").
			Scan(&res.Collections[i].FullNFTs).Error
		if err != nil {
			return
		}
	}
	return
}

func FNFTs(FNFTId string) (res []model.SNFT, err error) {
	err = DB.Where("LEFT(address, 40)=?", FNFTId).Find(&res).Error
	return
}
