package database

// NFTMeta NFT元信息
type NFTMeta struct {
	NFTAddr      string  `json:"nft_addr" gorm:"type:CHAR(42);primary_key"` //NFT地址
	Name         string  `json:"name"`                                      //名称
	Desc         string  `json:"desc"`                                      //描述
	Category     string  `json:"category"`                                  //分类
	SourceUrl    string  `json:"source_url"`                                //资源链接，图片或视频等文件链接
	CollectionId *string `json:"collection_id" gorm:"index"`                //所属合集id,合集名称+合集创建者+合集所在交易所的哈希
}

func (nm NFTMeta) Insert() error {
	return DB.Create(&nm).Error
}

// NFTAndMeta 用户NFT属性和元信息
type NFTAndMeta struct {
	UserNFT
	NFTMeta
}

func FetchUserNFTsAndMeta(exchanger, collectionId, owner string, page, size uint64) (data []NFTAndMeta, count int64, err error) {
	if exchanger != "" || collectionId != "" || owner != "" {
		where := ""
		if exchanger != "" {
			where += "exchanger_addr='" + exchanger + "'"
		}
		if collectionId != "" {
			if where != "" {
				where += " AND "
			}
			where += "collection_id='" + collectionId + "'"
		}
		if owner != "" {
			if where != "" {
				where += " AND "
			}
			where += "owner='" + owner + "'"
		}
		err = DB.Order("address DESC").Offset((page - 1) * size).Limit(size).
			Raw("SELECT * FROM user_nfts LEFT JOIN nft_meta ON user_nfts.address=nft_meta.nft_addr WHERE " + where).Scan(&data).Error
		if err != nil {
			return
		}
		err = DB.Where(where).Model(&UserNFT{}).Count(&count).Error
	} else {
		err = DB.Order("address DESC").Offset((page - 1) * size).Limit(size).
			Raw("SELECT * FROM user_nfts LEFT JOIN nft_meta ON user_nfts.address=nft_meta.nft_addr").Scan(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&UserNFT{}).Count(&count).Error
	}
	return
}

// SNFTAndMeta 用户NFT属性和元信息
type SNFTAndMeta struct {
	SNFT
	NFTMeta
}

func FetchSNFTsAndMeta(collectionId, owner string, page, size uint64) (data []SNFTAndMeta, count int64, err error) {
	if collectionId != "" || owner != "" {
		where := ""
		if collectionId != "" {
			where += "collection_id='" + collectionId + "'"
		}
		if owner != "" {
			if where != "" {
				where += " AND "
			}
			where += "owner='" + owner + "'"
		}
		err = DB.Order("create_number DESC").Offset((page - 1) * size).Limit(size).
			Raw("SELECT * FROM snfts LEFT JOIN nft_meta ON snfts.address=nft_meta.nft_addr WHERE " + where).Scan(&data).Error
		if err != nil {
			return
		}
		err = DB.Where(where).Model(&SNFT{}).Count(&count).Error
	} else {
		err = DB.Order("create_number DESC").Offset((page - 1) * size).Limit(size).
			Raw("SELECT * FROM snfts LEFT JOIN nft_meta ON snfts.address=nft_meta.nft_addr").Scan(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&SNFT{}).Count(&count).Error
	}
	return
}
