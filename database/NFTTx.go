package database

// NFTTx NFT交易属性信息
type NFTTx struct {
	//交易类型,1：转移、2:出价成交、3:定价购买、4：惰性定价购买、5：惰性定价购买、6：出价成交、7：惰性出价成交、8：撮合交易
	TxType        int32   `json:"tx_type"`
	NFTAddr       string  `json:"nft_addr" gorm:"type:CHAR(42);index"`      //交易的NFT地址
	ExchangerAddr *string `json:"exchanger_addr" gorm:"type:CHAR(42)"`      //交易所地址
	From          string  `json:"from" gorm:"type:CHAR(42);index"`          //卖家
	To            string  `json:"to" gorm:"type:CHAR(42);index"`            //买家
	Price         *string `json:"price"`                                    //价格,单位为wei
	Timestamp     uint64  `json:"timestamp"`                                //交易时间戳
	TxHash        string  `json:"tx_hash" gorm:"type:CHAR(66);primary_key"` //交易哈希
}

func (nt NFTTx) Insert() error {
	nft, err := FindUserNFT(nt.NFTAddr)
	if err != nil {
		return err
	}
	// 更新NFT所有者
	err = DB.Model(&UserNFT{}).Where("address=?", nft.Address).UpdateColumn(map[string]interface{}{
		"owner": nt.To,
	}).Error

	// 填充卖家字段（如果没有）
	if nt.From == "" {
		nt.From = nft.Owner
	}

	return DB.Create(&nt).Error
}

func FetchNFTTxs(address, exchanger, account string, page, size uint64) (data []NFTTx, count int64, err error) {
	if exchanger != "" || account != "" {
		where := ""
		if exchanger != "" {
			where += "exchanger_addr='" + exchanger + "'"
		}
		if account != "" {
			if where != "" {
				where += " AND "
			}
			where += "(`from`='" + account + "' OR `to`='" + account + "')"
		}
		if address != "" {
			if where != "" {
				where += " AND "
			}
			where += "nft_addr='" + address + "'"
		}
		err = DB.Where(where).Order("timestamp DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Where(where).Model(&NFTTx{}).Count(&count).Error
	} else {
		err = DB.Order("timestamp DESC").Offset(page - 1).Limit(size).Find(&data).Error
		if err != nil {
			return
		}
		err = DB.Model(&NFTTx{}).Count(&count).Error
	}
	return
}

func FindNFTTx(hash string) (data NFTTx, err error) {
	err = DB.Where("tx_hash=?", hash).First(&data).Error
	return
}
