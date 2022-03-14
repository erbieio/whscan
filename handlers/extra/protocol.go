package extra

// ErrRes 接口错误信息返回
type ErrRes struct {
	Err string `json:"err"` //错误信息
}

// CheckAuthRes 返回
type CheckAuthRes struct {
	Code   int64  `json:"code" `
	Msg    string `json:"msg"`
	Result uint64 `json:"result" ` //0 no exchange 1 no fee  2 normal
}

// CheckAuthReq 请求
type CheckAuthReq struct {
	Address string `form:"address" json:"address"` //区块号
}
