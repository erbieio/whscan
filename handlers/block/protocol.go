package block

import (
	"server/database"
)

// GetBlockRes 返回
type GetBlockRes struct {
	Code int64          `json:"code" `
	Msg  string         `json:"msg"`
	Data database.Block `json:"data" `
}

// GetBlockReq 请求
type GetBlockReq struct {
	BlockNumber string `form:"block_number" json:"block_number"` //区块号
}

// ErrRes 接口错误信息返回
type ErrRes struct {
	Err string `json:"err"` //错误信息
}

// GetTxRes 返回
type GetTxRes struct {
	Code int64                `json:"code" `
	Msg  string               `json:"msg"`
	Data database.Transaction `json:"data" `
}

// GetTxReq 请求
type GetTxReq struct {
	TxHash string `form:"tx_hash" json:"tx_hash"` //交易hash
}

// GetReceiptsRes 返回
type GetReceiptsRes struct {
	Code int64            `json:"code" `
	Msg  string           `json:"msg"`
	Data []database.TxLog `json:"data" `
}

// IndexReq 请求
type IndexReq struct {
}

// IndexRes 请求
type IndexRes struct {
	Code   int64   `json:"code" `
	Msg    string  `json:"msg"`
	Blocks []Block `json:"blocks"`
	Txs    []Tx    `json:"txs"`
}

type Block struct {
	Number   string `json:"number"`
	Ts       string `json:"timestamp"`
	Miner    string `json:"miner"`
	TxCount  int    `json:"tx_count"`
	GasUsed  string `json:"gas_used"`
	GasLimit string `json:"gas_limit"`
	Reword   string `json:"reword"`
}

type Tx struct {
	Hash   string `json:"hash"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Method string `json:"method"`
	Ts     string `json:"timestamp"`
}

// TxPageReq 请求
type TxPageReq struct {
	Page        int    `form:"page" json:"page"`                 //页
	PageSize    int    `form:"page_size" json:"page_size"`       //页码
	Address     string `form:"address" json:"address"`           //地址 不区分传 ""
	TxType      string `form:"tx_type" json:"tx_type"`           //交易类型  如ERC20  ERC1155 ERC721  internal  不区分传""
	BlockNumber string `form:"block_number" json:"block_number"` //区块号 不区分
}

// PageReq 请求
type PageReq struct {
	Page     int `form:"page" json:"page"`           //页
	PageSize int `form:"page_size" json:"page_size"` //页码
}

// ViewBlocksRes 请求
type ViewBlocksRes struct {
	Code  int64            `json:"code" `
	Msg   string           `json:"msg"`
	Data  []database.Block `json:"data"`
	Total int64            `json:"total" `
}

// ViewTxsRes 请求
type ViewTxsRes struct {
	Code  int64                  `json:"code" `
	Msg   string                 `json:"msg"`
	Data  []database.Transaction `json:"data"`
	Total int64                  `json:"total" `
}
