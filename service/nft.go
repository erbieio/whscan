package service

import "server/common/model"

// NFTsRes NFT paging return parameters
type NFTsRes struct {
	Total int64       `json:"total"` //The total number of NFTs
	NFTs  []model.NFT `json:"nfts"`  //NFT list
}

func FetchNFTs(owner string, page, size int) (res NFTsRes, err error) {
	db := DB.Model(&model.NFT{})
	if owner != "" {
		db = db.Where("owner=?", owner)
	}

	err = db.Count(&res.Total).Error
	if err != nil {
		return
	}
	err = db.Order("address DESC").Offset((page - 1) * size).Limit(size).Find(&res.NFTs).Error
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
	Creator      string `json:"creator"`      //creator address, also the address of royalty income
	MetaUrl      string `json:"meta_url"`     //Real meta information URL
	RoyaltyRatio uint32 `json:"royaltyRatio"` //the royalty rate of the same period of SNFT, the unit is one ten thousandth
	CreatedAt    int64  `json:"createdAt"`    //snft created time
}

func GetSNFT(addr string) (res SNFTRes, err error) {
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN epoches ON LEFT(address, 39) = epoches.id")
	err = db.Select("snfts.*, creator, timestamp AS created_at, meta_url, royalty_ratio").Where("address=?", addr).Scan(&res).Error
	return
}

// SNFTsAndMetaRes SNFT and meta information paging return parameters
type SNFTsAndMetaRes struct {
	Total int64     `json:"total"` //The total number of SNFTs
	NFTs  []SNFTRes `json:"nfts"`  //SNFT list
}

func FetchSNFTsAndMeta(owner string, page, size int) (res SNFTsAndMetaRes, err error) {
	db := DB.Model(&model.SNFT{}).Joins("LEFT JOIN epoches ON LEFT(address, 39) = epoches.id").Where("`remove`=false")
	if owner != "" {
		db = db.Where("owner=?", owner)
	}

	if owner == "" {
		res.Total = stats.TotalSNFT
	} else {
		if err = db.Count(&res.Total).Error; err != nil {
			return
		}
	}
	err = db.Select("snfts.*, creator, timestamp AS created_at, meta_url, royalty_ratio").Order("address DESC").Offset((page - 1) * size).Limit(size).Scan(&res.NFTs).Error
	return
}

func GetNFT(addr string) (res model.NFT, err error) {
	err = DB.Find(&res, "address=?", addr).Error
	return
}

func BlockSNFTs(number string) (res []model.SNFT, err error) {
	err = DB.Find(&res, "reward_number=?", number).Error
	return
}
