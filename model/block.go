package model

import "server/common/types"

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
	TxHash  types.Hash    `json:"txHash" gorm:"type:CHAR(66);index"` //所属交易哈希
	Address types.Address `json:"address" gorm:"type:CHAR(42)"`      //合约地址
	From    types.Address `json:"from" gorm:"type:CHAR(42);index"`   //发起地址
	To      types.Address `json:"to" gorm:"type:CHAR(42);index"`     //接收地址
	TokenId types.BigInt  `json:"tokenId" gorm:"type:VARCHAR(80)"`   //代币ID
	Value   types.BigInt  `json:"value" gorm:"type:VARCHAR(80)"`     //代币金额
}
