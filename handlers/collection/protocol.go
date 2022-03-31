package collection

import "server/database"

// ErrRes 错误返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误字符串
}

// PageRes NFT分页返回参数
type PageRes struct {
	Total       int64                 `json:"total"`       //NFT合集总数
	Collections []database.Collection `json:"collections"` //NFT合集列表
}
