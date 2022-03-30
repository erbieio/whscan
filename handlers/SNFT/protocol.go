package SNFT

import "server/database"

// ErrRes 错误返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误字符串
}

// PageRes SNFT分页返回参数
type PageRes struct {
	Total int64           `json:"total"` //SNFT总数
	NFTs  []database.SNFT `json:"nfts"`  //SNFT列表
}
