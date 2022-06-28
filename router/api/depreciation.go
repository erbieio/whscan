package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
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

// @Tags         obsolete interface
// @Summary      query exchange (new /exchanger/{addr})
// @Description  Query exchanges by address
// @Deprecated
// @Accept   json
// @Produce  json
// @Param    addr  query     string  true  "Exchange address"
// @Success  200   {object}  model.Exchanger
// @Failure  400   {object}  service.ErrRes
// @Router   /exchanger/get [get]
func _getExchanger(c *gin.Context) {
	getExchanger(c)
}

// requestErbTestReq request
type requestErbTestReq struct {
	Address string `form:"address" json:"address"` //Address
}

// requestErbTestRes returns
type requestErbTestRes struct {
	Code int64  `json:"code"` //0 success 1 wrong address other failure
	Msg  string `json:"msg"`
}

// @Tags         obsolete interface
// @Summary      request ERB test coin (new /erb_faucet)
// @Description  request ERB test coins
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

	err = service.SendErb(string(addr), context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, requestErbTestRes{Code: -1, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requestErbTestRes{Code: 0, Msg: "ok"})
}

// @Tags         obsolete interface
// @Summary      query block list (new /block/page)
// @Description  query block list
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

// @Tags         obsolete interface
// @Summary      query transaction list (new /transaction/page)
// @Description  query transaction list
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

// CheckAuthReq request
type CheckAuthReq struct {
	Address string `form:"address" json:"address"` //Address
}

// CheckAuthRes returns
type CheckAuthRes struct {
	Code int64    `json:"code"` //0 success 1 wrong address other failure
	Msg  string   `json:"msg"`
	Data *AuthRes `json:"data" `
}

// @Tags         obsolete interface
// @Summary      query exchange status (new /exchanger_auth)
// @Description  query exchange status
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

	status, flag, balance, err := service.ExchangerAuth(string(addr))
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

// @Tags         obsolete interface
// @Summary      query block (new /block/{number})
// @Description  query block
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

// @Tags         obsolete interface
// @Summary      query transaction (new /transaction/{hash})
// @Description  query transaction
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

// @Tags         obsolete interface
// @Summary      query receipt (new /transaction_logs/{hash})
// @Description  query receipt
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

// GetBlockRes returns
type GetBlockRes struct {
	Code int64         `json:"code" `
	Msg  string        `json:"msg"`
	Data service.Block `json:"data" `
}

// GetBlockReq request
type GetBlockReq struct {
	BlockNumber string `form:"block_number" json:"block_number"` //block number
}

// ErrRes interface error message returned
type ErrRes struct {
	Err string `json:"err"` //Error message
}

// GetTxRes returns
type GetTxRes struct {
	Code int64               `json:"code" `
	Msg  string              `json:"msg"`
	Data service.Transaction `json:"data" `
}

// GetTxReq request
type GetTxReq struct {
	TxHash string `form:"tx_hash" json:"tx_hash"` //Transaction hash
}

// GetReceiptsRes returns
type GetReceiptsRes struct {
	Code int64         `json:"code" `
	Msg  string        `json:"msg"`
	Data []service.Log `json:"data"`
}

// IndexReq request
type IndexReq struct {
}

// IndexRes request
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

// TxPageReq request
type TxPageReq struct {
	Page        int    `form:"page" json:"page"`                 //page
	PageSize    int    `form:"page_size" json:"page_size"`       //page number
	Address     string `form:"address" json:"address"`           //Address does not distinguish transmission
	BlockNumber string `form:"block_number" json:"block_number"` //Block number is indistinguishable
}

// PageReq request
type PageReq struct {
	Page     int `form:"page" json:"page"`           //page
	PageSize int `form:"page_size" json:"page_size"` //page number
}

// ViewBlocksRes request
type ViewBlocksRes struct {
	Code  int64           `json:"code" `
	Msg   string          `json:"msg"`
	Data  []service.Block `json:"data"`
	Total int64           `json:"total" `
}

// ViewTxsRes request
type ViewTxsRes struct {
	Code  int64                 `json:"code" `
	Msg   string                `json:"msg"`
	Data  []service.Transaction `json:"data"`
	Total int64                 `json:"total" `
}
