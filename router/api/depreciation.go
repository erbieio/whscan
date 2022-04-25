package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/extra"
	"server/service"
)

func Depreciation(e *gin.Engine) {
	e.GET("/exchanger/get", _getExchanger)
	e.GET("/extra/requestErbTest", requestErbTest)
	e.GET("/extra/checkAuth", checkAuth)
	e.GET("/block/getBlock", _getBlock)
	e.GET("/block/getTransaction", _getTransaction)
	e.GET("/block/getTransactionLogs", _getTransactionLogs)
	e.GET("/block/viewBlocks", viewBlocks)
	e.GET("/block/viewTransactions", viewTransactions)
}

// @Tags         过时接口
// @Summary      查询交易所(新/exchanger/{addr})
// @Description  按地址查询交易所
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    addr  query     string  true  "交易所地址"
// @Success  200   {object}  model.Exchanger
// @Failure  400   {object}  service.ErrRes
// @Router   /exchanger/get [get]
func _getExchanger(c *gin.Context) {
	getExchanger(c)
}

// requestErbTestReq 请求
type requestErbTestReq struct {
	Address string `form:"address" json:"address"` //地址
}

// requestErbTestRes  返回
type requestErbTestRes struct {
	Code int64  `json:"code"` //0 成功  1 地址有误 其他失败
	Msg  string `json:"msg"`
}

// @Tags         过时接口
// @Summary      请求ERB测试币(新/erb_faucet)
// @Description  请求ERB测试币
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     requestErbTestReq  true  "body"
// @Success  200   {object}  requestErbTestRes
// @Failure  400   {object}  requestErbTestRes
// @Router   /extra/requestErbTest [get]
func requestErbTest(c *gin.Context) {
	var req requestErbTestReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, requestErbTestRes{Code: -1, Msg: err.Error()})
		return
	}

	addr, err := utils.ParseAddress(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	err = extra.SendErb(string(addr), context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, requestErbTestRes{Code: -1, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requestErbTestRes{Code: 0, Msg: "ok"})
}

// @Tags         过时接口
// @Summary      查询区块列表（新/block/page）
// @Description  查询区块列表
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     PageReq  true  "body"
// @Success  200   {object}  ViewBlocksRes
// @Failure  400   {object}  ErrRes
// @Router   /block/viewBlocks [get]
func viewBlocks(c *gin.Context) {
	var req PageReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, IndexRes{Code: -1, Msg: err.Error()})
		return
	}
	blocks, count, err := service.FetchBlocks_(req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ViewBlocksRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ViewBlocksRes{Code: 0, Msg: "ok", Data: blocks, Total: count})
}

// @Tags         过时接口
// @Summary      查询交易列表（新/transaction/page）
// @Description  查询交易列表
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     TxPageReq  true  "body"
// @Success  200   {object}  ViewTxsRes
// @Failure  400   {object}  ErrRes
// @Router   /block/viewTransactions [get]
func viewTransactions(c *gin.Context) {
	var req TxPageReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ViewTxsRes{Code: -1, Msg: err.Error()})
		return
	}

	txs, count, err := service.FetchTxs(req.Page, req.PageSize, req.Address, req.BlockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
	}
	c.JSON(http.StatusOK, ViewTxsRes{Code: 0, Msg: "ok", Data: txs, Total: count})
}

// CheckAuthReq 请求
type CheckAuthReq struct {
	Address string `form:"address" json:"address"` //地址
}

// CheckAuthRes 返回
type CheckAuthRes struct {
	Code int64    `json:"code"` //0 成功  1 地址有误 其他失败
	Msg  string   `json:"msg"`
	Data *AuthRes `json:"data" `
}

// @Tags         过时接口
// @Summary      查询交易所状态(新/exchanger_auth)
// @Description  查询交易所状态
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     CheckAuthReq  true  "body"
// @Success  200   {object}  CheckAuthRes
// @Failure  400   {object}  ErrRes
// @Router   /extra/checkAuth [get]
func checkAuth(c *gin.Context) {
	var req CheckAuthReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}

	addr, err := utils.ParseAddress(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	status, flag, balance, err := extra.ExchangerAuth(string(addr))
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, CheckAuthRes{Code: 0, Msg: "ok", Data: &AuthRes{
		Status:           status,
		ExchangerFlag:    flag,
		ExchangerBalance: balance,
	}})
}

// @Tags         过时接口
// @Summary      查询区块（新/block/{number}）
// @Description  查询区块
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     GetBlockReq  true  "body"
// @Success  200   {object}  GetBlockRes
// @Failure  400   {object}  ErrRes
// @Router   /block/getBlock [get]
func _getBlock(c *gin.Context) {
	var req GetBlockReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
		return
	}
	block, err := service.FindBlock(req.BlockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetBlockRes{Code: 0, Msg: "ok", Data: block})
}

// @Tags         过时接口
// @Summary      查询交易（新/transaction/{hash}）
// @Description  查询交易
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     GetTxReq  true  "body"
// @Success  200   {object}  GetTxRes
// @Failure  400   {object}  ErrRes
// @Router   /block/getTransaction [get]
func _getTransaction(c *gin.Context) {
	var req GetTxReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetTxRes{Code: -1, Msg: err.Error()})
		return
	}
	tx, err := service.FindTx(req.TxHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetTxRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetTxRes{Code: 0, Msg: "ok", Data: tx})
}

// @Tags         过时接口
// @Summary      查询收据(新/transaction_logs/{hash})
// @Description  查询收据
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    body  query     GetTxReq  true  "body"
// @Success  200   {object}  GetReceiptsRes
// @Failure  400   {object}  ErrRes
// @Router   /block/getTransactionLogs [get]
func _getTransactionLogs(c *gin.Context) {
	var req GetTxReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetReceiptsRes{Code: -1, Msg: err.Error()})
		return
	}
	receipts, err := service.FindLogByTx(req.TxHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetReceiptsRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetReceiptsRes{Code: 0, Msg: "ok", Data: receipts})
}

// GetBlockRes 返回
type GetBlockRes struct {
	Code int64         `json:"code" `
	Msg  string        `json:"msg"`
	Data service.Block `json:"data" `
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
	Code int64               `json:"code" `
	Msg  string              `json:"msg"`
	Data service.Transaction `json:"data" `
}

// GetTxReq 请求
type GetTxReq struct {
	TxHash string `form:"tx_hash" json:"tx_hash"` //交易hash
}

// GetReceiptsRes 返回
type GetReceiptsRes struct {
	Code int64         `json:"code" `
	Msg  string        `json:"msg"`
	Data []service.Log `json:"data" `
}

// IndexReq 请求
type IndexReq struct {
}

// IndexRes 请求
type IndexRes struct {
	Code   int64   `json:"code" `
	Msg    string  `json:"msg"`
	Blocks []block `json:"blocks"`
	Txs    []Tx    `json:"txs"`
}

type block struct {
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
	Address     string `form:"address" json:"address"`           //地址 不区分传
	BlockNumber string `form:"block_number" json:"block_number"` //区块号 不区分
}

// PageReq 请求
type PageReq struct {
	Page     int `form:"page" json:"page"`           //页
	PageSize int `form:"page_size" json:"page_size"` //页码
}

// ViewBlocksRes 请求
type ViewBlocksRes struct {
	Code  int64           `json:"code" `
	Msg   string          `json:"msg"`
	Data  []service.Block `json:"data"`
	Total int64           `json:"total" `
}

// ViewTxsRes 请求
type ViewTxsRes struct {
	Code  int64                 `json:"code" `
	Msg   string                `json:"msg"`
	Data  []service.Transaction `json:"data"`
	Total int64                 `json:"total" `
}
