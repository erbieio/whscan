package extra

// ErrRes 接口错误信息返回
type ErrRes struct {
	Err string `json:"err"` //错误信息
}

// CheckAuthRes 返回
type CheckAuthRes struct {
	Code int64            `json:"code"` //0 成功  1 地址有误 其他失败
	Msg  string           `json:"msg"`
	Data CheckAuthResData `json:"data" `
}

type CheckAuthResData struct {
	Status           uint64 `json:"status"` //2 交易所付费状态正常  其他数字为欠费或者没交费
	ExchangerFlag    bool   `json:"exchanger_flag"`
	ExchangerBalance string `json:"exchanger_balance" `
}

// CheckAuthReq 请求
type CheckAuthReq struct {
	Address string `form:"address" json:"address"` //地址
}

// RequestErbTestReq 请求
type RequestErbTestReq struct {
	Address string `form:"address" json:"address"` //地址
}

// RequestErbTestRes  返回
type  RequestErbTestRes struct {
	Code int64            `json:"code"` //0 成功  1 地址有误 其他失败
	Msg  string           `json:"msg"`
}

