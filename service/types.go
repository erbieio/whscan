package service

const ZeroAddress = "0x0000000000000000000000000000000000000000"

// ErrRes 接口错误信息返回
type ErrRes struct {
	ErrStr string `json:"err_str"` //错误信息
}
