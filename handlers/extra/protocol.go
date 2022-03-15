package extra

// ErrRes 接口错误信息返回
type ErrRes struct {
	Err string `json:"err"` //错误信息
}

// CheckAuthRes 返回
type CheckAuthRes struct {
	Code   int64  `json:"code"` //0 成功  其他失败
	Msg    string `json:"msg"`
	Result uint64 `json:"result" ` //0 没有创建过交易所 1 欠费  2 交易所状态正常
}

// CheckAuthReq 请求
type CheckAuthReq struct {
	Address string `form:"address" json:"address"` //地址
}
