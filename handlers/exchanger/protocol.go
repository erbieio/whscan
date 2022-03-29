package exchanger

import "server/database"

// ErrRes 错误返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误字符串
}

// PageRes 交易所分页返回参数
type PageRes struct {
	Total      int64                `json:"total"`      //交易所总数
	Exchangers []database.Exchanger `json:"exchangers"` //交易所列表
}
