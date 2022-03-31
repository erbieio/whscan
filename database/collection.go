package database

// Collection 合集信息
type Collection struct {
	Id          string `json:"id" gorm:"type:CHAR(66);primary_key"`  //名称+创建者+所属交易所的哈希
	Name        string `json:"name"`                                 //名称，唯一标识合集
	Creator     string `json:"creator" gorm:"index"`                 //创建者，唯一标识合集
	Category    string `json:"category"`                             //分类
	Desc        string `json:"desc"`                                 //描述
	ImgUrl      string `json:"img_url"`                              //图片链接
	BlockNumber uint64 `json:"block_number" gorm:"index"`            //创建区块高度，等于合集第一个NFT的
	Exchanger   string `json:"exchanger" gorm:"type:CHAR(42);index"` //所属交易所，唯一标识合集
}

func (c Collection) Save() {
	DB.Create(&c)
}

func FetchCollections(exchanger, creator string, page, size uint64) (data []Collection, count int64, err error) {
	if exchanger != "" || creator != "" {
		where := ""
		if exchanger != "" {
			where += "exchanger_addr='" + exchanger + "'"
		}
		if creator != "" {
			if exchanger != "" {
				where += " AND "
			}
			where += "creator='" + creator + "'"
		}
		err = DB.Where(where).Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Where(where).Model(&Collection{}).Count(&count).Error
	} else {
		err = DB.Order("block_number DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&Collection{}).Count(&count).Error
	}
	return
}

func GetCollection(id string) (c Collection, err error) {
	err = DB.Where("id=?", id).Find(&c).Error
	return
}
