// Package model 数据库表定义
package model

import (
	"gorm.io/gorm"
	"server/common/types"
)

var Tables = []interface{}{
	&Block{},
	&Uncle{},
	&Transaction{},
	&Log{},
	&Account{},
	&Contract{},
	&InternalTx{},
	&ERC20Transfer{},
	&ERC721Transfer{},
	&ERC1155Transfer{},
	&Exchanger{},
	&UserNFT{},
	&OfficialNFT{},
	&NFTMeta{},
	&Collection{},
	&NFTTx{},
	&ConsensusPledge{},
	&ExchangerPledge{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Tables...)
}

func DropTable(db *gorm.DB) error {
	return db.Migrator().DropTable(Tables...)
}

// Header 区块头信息
type Header struct {
	Difficulty       types.Uint64   `json:"difficulty"`                            //难度
	ExtraData        string         `json:"extraData"`                             //额外数据
	GasLimit         types.Uint64   `json:"gasLimit"`                              //燃料上限
	GasUsed          types.Uint64   `json:"gasUsed"`                               //燃料消耗
	Hash             types.Hash     `json:"hash" gorm:"type:CHAR(66);primaryKey"`  //哈希
	Miner            types.Address  `json:"miner" gorm:"type:CHAR(42)"`            //矿工
	MixHash          types.Hash     `json:"mixHash" gorm:"type:CHAR(66)"`          //混合哈希
	Nonce            types.Bytes8   `json:"nonce" gorm:"type:CHAR(18)"`            //难度随机数
	ParentHash       types.Hash     `json:"parentHash" gorm:"type:CHAR(66)"`       //父区块哈希
	ReceiptsRoot     types.Hash     `json:"receiptsRoot" gorm:"type:CHAR(66)"`     //交易收据根哈希
	Sha3Uncles       types.Hash     `json:"sha3Uncles" gorm:"type:CHAR(66)"`       //叔块根哈希
	Size             types.Uint64   `json:"size"`                                  //大小
	StateRoot        types.Hash     `json:"stateRoot" gorm:"type:CHAR(66)"`        //世界树根哈希
	Timestamp        types.Uint64   `json:"timestamp"`                             //时间戳
	TransactionsRoot types.Hash     `json:"transactionsRoot" gorm:"type:CHAR(66)"` //交易根哈希
	UncleHashes      types.StrArray `json:"uncles" gorm:"type:TEXT"`               //叔块哈希
	UnclesCount      types.Uint64   `json:"unclesCount"`                           //叔块数量
}

// Uncle 叔块信息
type Uncle struct {
	Header
	Number types.Uint64 `json:"number"` //叔块号
}

// Block 区块信息
type Block struct {
	Header
	Number           types.Uint64 `json:"number" gorm:"uniqueIndex"`               //区块号
	TotalDifficulty  types.BigInt `json:"totalDifficulty" gorm:"type:VARCHAR(64)"` //总难度
	TotalTransaction types.Uint64 `json:"totalTransaction"`                        //交易数量
}

// Account 账户信息
type Account struct {
	Name     *string       `json:"name" gorm:"type:VARCHAR(64)"`            //名称
	Address  types.Address `json:"address" gorm:"type:CHAR(42);primaryKey"` //地址
	Balance  types.BigInt  `json:"balance" gorm:"type:VARCHAR(128);index"`  //余额
	Nonce    types.Uint64  `json:"transactionCount"`                        //交易随机数，交易量
	CodeHash *types.Hash   `json:"codeHash" gorm:"type:CHAR(66)"`           //合约字节码哈希，普通账户为空
}

// Contract 合约信息
type Contract struct {
	Address   types.Address `json:"address" gorm:"type:CHAR(42);primaryKey"` //合约地址
	Code      string        `json:"code"`                                    //字节码
	ERC       types.ERC     `json:"ERC"`                                     //ERC类型，ERC20,ERC721,ERC1155
	Creator   types.Address `json:"creator" gorm:"type:CHAR(42)"`            //创建者
	CreatedTx types.Hash    `json:"createdTx" gorm:"type:CHAR(66)"`          //创建交易
}

// Transaction 交易信息
type Transaction struct {
	BlockHash   types.Hash     `json:"blockHash" gorm:"type:CHAR(66)"`          //区块哈希
	BlockNumber types.Uint64   `json:"blockNumber" gorm:"index:,sort:DESC"`     //区块号
	From        types.Address  `json:"from" gorm:"type:CHAR(42);index"`         //发送地址
	To          *types.Address `json:"to" gorm:"type:CHAR(42);index"`           //接收地址
	Input       string         `json:"input"`                                   //额外输入数据，合约调用编码数据
	MethodId    *types.Bytes8  `json:"methodId,omitempty" gorm:"type:CHAR(18)"` //方法ID，普通交易为空
	Value       types.BigInt   `json:"value" gorm:"type:VARCHAR(128)"`          //金额，单位wei
	Nonce       types.Uint64   `json:"nonce"`                                   //随机数，发起账户的交易次数
	Gas         types.Uint64   `json:"gas"`                                     //燃料
	GasPrice    types.Uint64   `json:"gasPrice"`                                //燃料价格
	Hash        types.Hash     `json:"hash" gorm:"type:CHAR(66);primaryKey"`    //哈希
	Receipt
}

// Receipt 交易收据
type Receipt struct {
	Status            *types.Uint64  `json:"status,omitempty"`                         //状态，1：成功；0：失败
	CumulativeGasUsed types.Uint64   `json:"cumulativeGasUsed"`                        //累计燃料消耗
	ContractAddress   *types.Address `json:"contractAddress" gorm:"type:CHAR(42)"`     //创建的合约地址
	GasUsed           types.Uint64   `json:"gasUsed"`                                  //燃料消耗
	TxIndex           types.Uint64   `json:"transactionIndex" gorm:"index:,sort:DESC"` //在区块内的序号
}

// Log 交易日志
type Log struct {
	Address types.Address  `json:"address" gorm:"type:CHAR(42)"`                          //所属合约地址
	EventID *types.Hash    `json:"eventID" gorm:"type:CHAR(66)"`                          //事件ID
	Topics  types.StrArray `json:"topics" gorm:"type:VARCHAR(277)"`                       //主题
	Data    string         `json:"data"`                                                  //数据
	Removed bool           `json:"removed"`                                               //是否移除
	TxHash  types.Hash     `json:"transactionHash" gorm:"type:CHAR(66);primaryKey;index"` //所属交易哈希
	Index   types.Uint64   `json:"logIndex" gorm:"primaryKey"`                            //在交易内的序号
}

// InternalTx 内部交易
type InternalTx struct {
	ParentTxHash types.Hash     `json:"parentTxHash" gorm:"type:CHAR(66);index"` //所属交易
	Depth        types.Uint64   `json:"depth"`                                   //调用深度
	Op           string         `json:"op" gorm:"type:VARCHAR(16)"`              //操作
	From         *types.Address `json:"from" gorm:"type:CHAR(42);index"`         //发起地址
	To           *types.Address `json:"to" gorm:"type:CHAR(42);index"`           //接收地址
	Value        types.BigInt   `json:"value" gorm:"type:VARCHAR(128)"`          //金额，单位wei
	GasLimit     types.Uint64   `json:"gasLimit"`                                //燃料上限
}

// ERC20Transfer ERC20合约转移事件
type ERC20Transfer struct {
	TxHash  types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"`  //所属交易哈希
	Address types.Address `json:"address" gorm:"type:CHAR(42);index"` //合约地址
	From    types.Address `json:"from" gorm:"type:CHAR(42);index"`    //发起地址
	To      types.Address `json:"to" gorm:"type:CHAR(42);index"`      //接收地址
	Value   types.BigInt  `json:"value" gorm:"type:VARCHAR(80)"`      //金额
}

// ERC721Transfer ERC721合约转移事件
type ERC721Transfer struct {
	TxHash  types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //所属交易哈希
	Address types.Address `json:"address" gorm:"type:CHAR(42)"`      //合约地址
	From    types.Address `json:"from" gorm:"type:CHAR(42);index"`   //发起地址
	To      types.Address `json:"to" gorm:"type:CHAR(42);index"`     //接收地址
	TokenId types.BigInt  `json:"tokenId" gorm:"type:VARCHAR(80)"`   //代币ID
}

// ERC1155Transfer ERC1155合约转移事件
type ERC1155Transfer struct {
	TxHash   types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //所属交易哈希
	Address  types.Address `json:"address" gorm:"type:CHAR(42)"`      //合约地址
	Operator types.Address `json:"operator" gorm:"type:CHAR(42)"`     //操作者地址
	From     types.Address `json:"from" gorm:"type:CHAR(42);index"`   //发起地址
	To       types.Address `json:"to" gorm:"type:CHAR(42);index"`     //接收地址
	TokenId  types.BigInt  `json:"tokenId" gorm:"type:VARCHAR(80)"`   //代币ID
	Value    types.BigInt  `json:"value" gorm:"type:VARCHAR(80)"`     //代币金额
}

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

// Pledge 账户质押金额
type Pledge struct {
	Address string `json:"address" gorm:"type:CHAR(44);primary_key"` //质押账户
	Amount  string `json:"amount" gorm:"type:CHAR(64)"`              //质押金额
	Count   uint64 `json:"count"`                                    //质押次数，PledgeAdd和PledgeSub都加一次
}

// ExchangerPledge 交易所质押
type ExchangerPledge Pledge

// ConsensusPledge 共识质押
type ConsensusPledge Pledge