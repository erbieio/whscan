package NFT

import "server/database"

// ErrRes 错误返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误字符串
}

// PageRes NFT分页返回参数
type PageRes struct {
	Total int64              `json:"total"` //NFT总数
	NFTs  []database.UserNFT `json:"nfts"`  //NFT列表
}

// PageMetaRes NFT和元信息分页返回参数
type PageMetaRes struct {
	Total int64                 `json:"total"` //NFT总数
	NFTs  []database.NFTAndMeta `json:"nfts"`  //NFT列表
}

// PageTxRes NFT交易分页返回参数
type PageTxRes struct {
	Total  int64            `json:"total"`   //NFT总数
	NFTTxs []database.NFTTx `json:"nft_txs"` //NFT交易列表
}
