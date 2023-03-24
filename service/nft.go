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
	db := DB.Model(&model.NFT{}).Joins("LEFT JOIN collections ON id=collection_id")
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

func EpochNFTTxs(epoch string, page, size int) (res NFTTxsRes, err error) {
	db := DB.Model(&model.NFTTx{}).Where("LEFT(nft_addr,39)=?", epoch)
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	err = db.Order("timestamp DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTTxs).Error
	return
}

func GetNFTTx(hash string) (res model.NFTTx, err error) {
	err = DB.Where("tx_hash=?", hash).Take(&res).Error
	return
}

// SNFTsRes SNFT paging return parameters
type SNFTsRes struct {
	Total int64        `json:"total"` //The total number of SNFTs
	NFTs  []model.SNFT `json:"nfts"`  //SNFT list
}

func FetchSNFTs(sort, owner string, page, size int) (res SNFTsRes, err error) {
	db := DB.Model(&model.SNFT{}).Where("`remove`=false")
	if owner == "" {
		res.Total = stats.TotalSNFT
	} else {
		db = db.Where("`owner`=?", owner)
		if err = db.Count(&res.Total).Error; err != nil {
			return
		}
	}
	if sort == "1" {
		db = db.Order("LENGTH(`address`) ASC,`address` DESC")
	} else {
		db = db.Order("`address` DESC")
	}
	err = db.Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

type SNFTRes struct {
	model.SNFT
	model.FNFT
	Creator        string `json:"creator"`        //creator address, also the address of royalty income
	RoyaltyRatio   uint32 `json:"royaltyRatio"`   //the royalty rate of the same period of SNFT, the unit is one ten thousandth
	Exchanger      string `json:"exchanger"`      //exchanger address
	CollectionName string `json:"collectionName"` //collection name
	CreatedAt      int64  `json:"createdAt"`      //snft created time
}

func GetSNFT(addr string) (res SNFTRes, err error) {
	db := DB.Model(&model.SNFT{}).
		Joins("LEFT JOIN fnfts ON LEFT(address, 41) = fnfts.id").
		Joins("LEFT JOIN collections ON LEFT(address, 40) = collections.id").
		Joins("LEFT JOIN epoches ON LEFT(address, 39) = epoches.id")
	err = db.Select("snfts.*, fnfts.*, epoches.creator, epoches.timestamp AS created_at, royalty_ratio,collections.name AS collection_name").Where("address=?", addr).Scan(&res).Error
	return
}

// SNFTsAndMetaRes SNFT and meta information paging return parameters
type SNFTsAndMetaRes struct {
	Total int64      `json:"total"` //The total number of SNFTs
	NFTs  []*SNFTRes `json:"nfts"`  //SNFT list
}

func FetchSNFTsAndMeta(owner, collectionId string, page, size int) (res SNFTsAndMetaRes, err error) {
	db := DB.Model(&model.SNFT{}).
		Joins("LEFT JOIN fnfts ON LEFT(address, 41) = fnfts.id").
		Joins("LEFT JOIN collections ON LEFT(address, 40) = collections.id").
		Joins("LEFT JOIN epoches ON LEFT(address, 39) = epoches.id").
		Where("`remove`=false")
	if owner != "" {
		db = db.Where("owner=?", owner)
	}
	if collectionId != "" {
		db = db.Where("collection.id=?", collectionId)
	}
	if owner == "" && collectionId == "" {
		res.Total = stats.TotalSNFT
	} else {
		if err = db.Count(&res.Total).Error; err != nil {
			return
		}
	}
	err = db.Select("snfts.*, fnfts.*, epoches.creator, epoches.timestamp AS created_at, royalty_ratio,collections.name AS collection_name").Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

type NFTRes struct {
	model.NFT
	CollectionName string `json:"collectionName"`
}

func GetNFT(addr string) (res NFTRes, err error) {
	err = DB.Model(&model.NFT{}).Joins("LEFT JOIN collections ON collection_id=collections.id").
		Where("address=?", addr).Select("nfts.*,collections.name AS collection_name").Scan(&res).Error
	return
}

func GetRecycleTx(hash, addr string) (res *model.NFTTx, err error) {
	err = DB.Model(&model.NFTTx{}).Where("tx_type=6 AND (tx_hash=? OR nft_addr=?)", hash, addr).Limit(1).Scan(&res).Error
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
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN collections on LEFT(address,40) = id AND owner=?", owner)
	err = db.Select("`collections`.*,COUNT(address) AS total_hold").Group("id").
		Order("id DESC").Offset((page - 1) * size).Limit(size).Scan(&res.Collections).Error
	if err != nil {
		return
	}
	if err = db.Count(&res.Total).Error; err != nil {
		return
	}
	for i := range res.Collections {
		db = DB.Model(&model.SNFT{}).Joins("LEFT JOIN fnfts on LEFT(address,41) = fnfts.id").Select("fnfts.*,COUNT(*) AS total_hold")
		err = db.Where("LEFT(address, 40)=? AND owner=?", res.Collections[i].Id, owner).Group("id").
			Scan(&res.Collections[i].FullNFTs).Error
		if err != nil {
			return
		}
	}
	return
}

func FNFTs(FNFTId string) (res []model.SNFT, err error) {
	err = DB.Where("LEFT(address, 41)=?", FNFTId).Find(&res).Error
	return
}
