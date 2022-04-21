package model

// UserNFT 用户NFT属性信息
type UserNFT struct {
	Address       *string `json:"address" gorm:"type:CHAR(44);primary_key"`  //NFT地址
	RoyaltyRatio  uint32  `json:"royalty_ratio"`                             //版税费率,单位万分之一
	MetaUrl       string  `json:"meta_url"`                                  //元信息URL
	ExchangerAddr string  `json:"exchanger_addr" gorm:"type:CHAR(44);index"` //所在交易所地址,没有的可以在任意交易所交易
	LastPrice     *string `json:"last_price"`                                //最后成交价格(未成交为null)，单位wei
	Creator       string  `json:"creator" gorm:"type:CHAR(44)"`              //创建者地址
	Timestamp     uint64  `json:"timestamp"`                                 //创建时间戳
	BlockNumber   uint64  `json:"block_number"`                              //创建的区块高度
	TxHash        string  `json:"tx_hash" gorm:"type:CHAR(66)"`              //创建的交易哈希
	Owner         string  `json:"owner" gorm:"type:CHAR(44);index"`          //所有者
}

// OfficialNFT OfficialNFT属性信息
type OfficialNFT struct {
	Address      string  `json:"address" gorm:"type:CHAR(44);primary_key"` //OfficialNFT地址
	CreateAt     uint64  `json:"create_at"`                                //创建时间戳
	CreateNumber uint64  `json:"create_number" gorm:"index"`               //创建的区块高度
	Creator      string  `json:"creator" gorm:"type:CHAR(44)"`             //创建者地址
	LastPrice    *string `json:"last_price"`                               //最后成交价格(未成交为null)，单位wei
	Awardee      *string `json:"awardee"`                                  //被奖励的矿工地址
	RewardAt     *uint64 `json:"reward_at"`                                //奖励时间戳,矿工被奖励这个OfficialNFT的时间
	RewardNumber *uint64 `json:"reward_number"`                            //奖励区块高度,矿工被奖励这个OfficialNFT的区块高度
	Owner        *string `json:"owner" gorm:"type:CHAR(44)"`               //所有者,未分配和回收的为null
	RoyaltyRatio uint32  `json:"royalty_ratio"`                            //版税费率,单位万分之一
	MetaUrl      string  `json:"meta_url"`                                 //元信息链接
}

// NFTTx NFT交易属性信息
type NFTTx struct {
	//交易类型,1：转移、2:出价成交、3:定价购买、4：惰性定价购买、5：惰性定价购买、6：出价成交、7：惰性出价成交、8：撮合交易
	TxType        int32   `json:"tx_type"`
	NFTAddr       *string `json:"nft_addr" gorm:"type:CHAR(42);index"`      //交易的NFT地址
	ExchangerAddr *string `json:"exchanger_addr" gorm:"type:CHAR(42)"`      //交易所地址
	From          string  `json:"from" gorm:"type:CHAR(42);index"`          //卖家
	To            string  `json:"to" gorm:"type:CHAR(42);index"`            //买家
	Price         *string `json:"price"`                                    //价格,单位为wei
	Timestamp     uint64  `json:"timestamp"`                                //交易时间戳
	TxHash        string  `json:"tx_hash" gorm:"type:CHAR(66);primary_key"` //交易哈希
}

// NFTMeta NFT元信息
type NFTMeta struct {
	NFTAddr      string  `json:"nft_addr" gorm:"type:CHAR(42);primary_key"` //NFT地址
	Name         string  `json:"name"`                                      //名称
	Desc         string  `json:"desc"`                                      //描述
	Category     string  `json:"category"`                                  //分类
	SourceUrl    string  `json:"source_url"`                                //资源链接，图片或视频等文件链接
	CollectionId *string `json:"collection_id" gorm:"index"`                //所属合集id,合集名称+合集创建者+合集所在交易所的哈希
}

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

// Exchanger 交易所属性信息
type Exchanger struct {
	Address     string `json:"address" gorm:"type:CHAR(42);primary_key"` //交易所地址
	Name        string `json:"name" gorm:"type:VARCHAR(256)"`            //交易所名称
	URL         string `json:"url"`                                      //交易所URL
	FeeRatio    uint32 `json:"fee_ratio"`                                //手续费率,单位万分之一
	Creator     string `json:"creator" gorm:"type:CHAR(42)"`             //创建者地址
	Timestamp   uint64 `json:"timestamp" gorm:"index"`                   //开启时间
	IsOpen      bool   `json:"is_open"`                                  //是否开启中
	BlockNumber uint64 `json:"block_number" gorm:"index"`                //创建时的区块号
	TxHash      string `json:"tx_hash" gorm:"type:CHAR(66)"`             //创建的交易
	NFTCount    uint64 `json:"nft_count" gorm:"-"`                       //NFT总数，批量查询的此字段无效
}
