package model

import "server/common/types"

type Header struct {
	Difficulty       types.Uint64   `json:"difficulty"`
	ExtraData        string         `json:"extraData"`
	GasLimit         types.Uint64   `json:"gasLimit"`
	GasUsed          types.Uint64   `json:"gasUsed"`
	Hash             types.Bytes32  `json:"hash" gorm:"type:CHAR(66);primaryKey"`
	Miner            types.Bytes20  `json:"miner" gorm:"type:CHAR(42)"`
	MixHash          types.Bytes32  `json:"mixHash" gorm:"type:CHAR(66)"`
	Nonce            types.Bytes8   `json:"nonce" gorm:"type:CHAR(18)"`
	ParentHash       types.Bytes32  `json:"parentHash" gorm:"type:CHAR(66)"`
	ReceiptsRoot     types.Bytes32  `json:"receiptsRoot" gorm:"type:CHAR(66)"`
	Sha3Uncles       types.Bytes32  `json:"sha3Uncles" gorm:"type:CHAR(66)"`
	Size             types.Uint64   `json:"size"`
	StateRoot        types.Bytes32  `json:"stateRoot" gorm:"type:CHAR(66)"`
	Timestamp        types.Uint64   `json:"timestamp"`
	TransactionsRoot types.Bytes32  `json:"transactionsRoot" gorm:"type:CHAR(66)"`
	UncleHashes      types.StrArray `json:"uncles" gorm:"type:TEXT"`
	UnclesCount      types.Uint64   `json:"unclesCount"`
}

type Uncle struct {
	Header
	Number types.Uint64
}

type Block struct {
	Header
	Number           types.Uint64   `json:"number" gorm:"uniqueIndex"`
	TotalDifficulty  string         `json:"totalDifficulty" gorm:"type:VARCHAR(64)"`
	TotalTransaction types.Uint64   `json:"totalTransaction"`
	Transactions     types.StrArray `json:"transactions"`
}

type Account struct {
	Name     *string        `gorm:"type:VARCHAR(64)"`
	Address  types.Bytes20  `gorm:"type:CHAR(42);primaryKey"`
	Balance  string         `gorm:"type:VARCHAR(128)"`
	Nonce    types.Uint64   `json:"transactionCount"`
	CodeHash *types.Bytes32 `gorm:"type:CHAR(66)"`
}

type Contract struct {
	Address   types.Bytes20 `gorm:"type:CHAR(42);primaryKey"`
	Code      string
	ERC       types.ERC
	Creator   types.Bytes20 `gorm:"type:CHAR(42)"`
	CreatedTx types.Bytes32 `gorm:"type:CHAR(66)"`
}

type Transaction struct {
	BlockHash   types.Bytes32  `json:"blockHash" gorm:"type:CHAR(66)"`
	BlockNumber types.Uint64   `json:"blockNumber" gorm:"index:,sort:DESC"`
	From        types.Bytes20  `json:"from" gorm:"type:CHAR(42);index"`
	To          *types.Bytes20 `json:"to" gorm:"type:CHAR(42);index"`
	Input       string         `json:"input"`
	MethodId    *types.Bytes8  `json:"methodId,omitempty" gorm:"type:CHAR(18)"`
	Value       string         `json:"value" gorm:"type:VARCHAR(128)"`
	Nonce       types.Uint64   `json:"nonce"`
	Gas         types.Uint64   `json:"gas"`
	GasPrice    types.Uint64   `json:"gasPrice"`
	Hash        types.Bytes32  `json:"hash" gorm:"type:CHAR(66);primaryKey"`
	Receipt
}

type Receipt struct {
	Status            *types.Uint64  `json:"status,omitempty"`
	CumulativeGasUsed types.Uint64   `json:"cumulativeGasUsed"`
	ContractAddress   *types.Bytes20 `json:"contractAddress" gorm:"type:CHAR(42)"`
	EffectiveGasPrice types.Uint64   `json:"effectiveGasPrice"`
	GasUsed           types.Uint64   `json:"gasUsed"`
	TxIndex           types.Uint64   `json:"transactionIndex" gorm:"index:,sort:DESC"`
}

type Log struct {
	Address types.Bytes20  `json:"address" gorm:"type:CHAR(42)"`
	EventID *types.Bytes32 `json:"eventID" gorm:"type:CHAR(66)"`
	Topics  types.StrArray `json:"topics" gorm:"type:VARCHAR(277)"`
	Data    string         `json:"data"`
	Removed bool           `json:"removed"`
	TxHash  types.Bytes32  `json:"transactionHash" gorm:"type:CHAR(66);primaryKey;index"`
	Index   types.Uint64   `json:"logIndex" gorm:"primaryKey"`
}

type InternalTx struct {
	ParentTxHash types.Bytes32  `json:"parentTxHash" gorm:"type:CHAR(66);index"`
	Depth        types.Uint64   `json:"depth"`
	Op           string         `json:"op" gorm:"type:VARCHAR(16)"`
	From         *types.Bytes20 `json:"from" gorm:"type:CHAR(42);index"`
	To           *types.Bytes20 `json:"to" gorm:"type:CHAR(42);index"`
	Value        string         `json:"value" gorm:"type:VARCHAR(128)"`
	GasLimit     types.Uint64   `json:"gasLimit"`
}

type ERC20Transfer struct {
	TxHash  types.Bytes32 `gorm:"type:CHAR(66);index"`
	Address types.Bytes20 `gorm:"type:CHAR(42);index"`
	From    types.Bytes20 `json:"from" gorm:"type:CHAR(42);index"`
	To      types.Bytes20 `json:"to" gorm:"type:CHAR(42);index"`
	Value   string        `gorm:"type:VARCHAR(66)"`
}

type ERC721Transfer struct {
	TxHash  types.Bytes32 `gorm:"type:CHAR(66);index"`
	Address types.Bytes20 `gorm:"type:CHAR(42)"`
	From    types.Bytes20 `json:"from" gorm:"type:CHAR(42);index"`
	To      types.Bytes20 `json:"to" gorm:"type:CHAR(42);index"`
	TokenId string        `gorm:"type:VARCHAR(66)"`
}

type ERC1155Transfer struct {
	TxHash  types.Bytes32 `gorm:"type:CHAR(66);index"`
	Address types.Bytes20 `gorm:"type:CHAR(42)"`
	From    types.Bytes20 `json:"from" gorm:"type:CHAR(42);index"`
	To      types.Bytes20 `json:"to" gorm:"type:CHAR(42);index"`
	TokenId string        `gorm:"type:VARCHAR(66)"`
	Value   string        `gorm:"type:VARCHAR(66)"`
}
