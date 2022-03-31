package block

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/database"
)

func Routers(e *gin.Engine) {
	e.GET("/block/getBlock", getBlock)
	e.GET("/block/getTransaction", getTransaction)
	e.GET("/block/getTransactionLogs", getTransactionLogs)
	e.GET("/block/viewBlocks", viewBlocks)
	e.GET("/block/viewTransactions", viewTransactions)
}

// @Tags  区块查询
// @Summary 查询区块列表
// @Description 查询区块列表
// @Accept json
// @Produce json
// @Param body body PageReq true "body"
// @Success 200 {object} ViewBlocksRes
// @Failure 400 {object} ErrRes
// @Router /block/viewBlocks [get]
func viewBlocks(c *gin.Context) {
	var req PageReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, IndexRes{Code: -1, Msg: err.Error()})
		return
	}
	blocks, count, err := database.FetchBlocks(req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ViewBlocksRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ViewBlocksRes{Code: 0, Msg: "ok", Data: blocks, Total: count})
}

// @Tags  区块查询
// @Summary 查询交易列表
// @Description 查询交易列表
// @Accept json
// @Produce json
// @Param body body TxPageReq true "body"
// @Success 200 {object} ViewTxsRes
// @Failure 400 {object} ErrRes
// @Router /block/viewTransactions [get]
func viewTransactions(c *gin.Context) {
	var req TxPageReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ViewTxsRes{Code: -1, Msg: err.Error()})
		return
	}

	txs, count, err := database.FetchTxs(req.Page, req.PageSize, req.Address, req.TxType, req.BlockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
	}
	c.JSON(http.StatusOK, ViewTxsRes{Code: 0, Msg: "ok", Data: txs, Total: count})
}

// @Tags  区块查询
// @Summary 查询区块
// @Description 查询区块
// @Accept json
// @Produce json
// @Param body body GetBlockReq true "body"
// @Success 200 {object} GetBlockRes
// @Failure 400 {object} ErrRes
// @Router /block/getBlock [get]
func getBlock(c *gin.Context) {
	var req GetBlockReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
		return
	}
	block, err := database.FindBlock(req.BlockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetBlockRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetBlockRes{Code: 0, Msg: "ok", Data: block})
}

// @Tags  区块查询
// @Summary 查询交易
// @Description 查询交易
// @Accept json
// @Produce json
// @Param body body GetTxReq true "body"
// @Success 200 {object} GetTxRes
// @Failure 400 {object} ErrRes
// @Router /block/getTransaction [get]
func getTransaction(c *gin.Context) {
	var req GetTxReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetTxRes{Code: -1, Msg: err.Error()})
		return
	}
	tx, err := database.FindTx(req.TxHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetTxRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetTxRes{Code: 0, Msg: "ok", Data: tx})
}

// @Tags  区块查询
// @Summary 查询收据
// @Description 查询收据
// @Accept json
// @Produce json
// @Param body body GetTxReq true "body"
// @Success 200 {object} GetReceiptsRes
// @Failure 400 {object} ErrRes
// @Router /block/getTransactionLogs [get]
func getTransactionLogs(c *gin.Context) {
	var req GetTxReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetReceiptsRes{Code: -1, Msg: err.Error()})
		return
	}
	receipts, err := database.FindLogByTx(req.TxHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetReceiptsRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, GetReceiptsRes{Code: 0, Msg: "ok", Data: receipts})
}
