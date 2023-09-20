package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Transaction transactionAPI
func Transaction(e *gin.Engine) {
	e.GET("/transaction/page", pageTransaction)
	e.GET("/transaction/:hash", getTransaction)
	e.GET("/transaction_logs/:hash", getTransactionLogs)
	e.GET("/transaction/internal/page", pageInternalTransaction)
	e.GET("/transaction/internal/:hash", getInternalTransaction)
	e.GET("/transaction/erbie/page", pageErbieTransaction)
	e.GET("/transaction/erbie/:hash", getErbieTransaction)
}

// @Tags        transaction
// @Summary     query transaction list
// @Description query transaction list in reverse order
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Param       number    query    string false "Block number, if empty, query all"
// @Param       addr      query    string false "Account address, if empty, query all"
// @Param       types     query    string false "erbie tx type,supports multiple types,if empty, query all"
// @Success     200       {object} service.TransactionsRes
// @Failure     400       {object} service.ErrRes
// @Router      /transaction/page [get]
func pageTransaction(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.FetchTransactions(page, size, c.Query("number"), c.Query("addr"), c.Query("types"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query transaction
// @Description specifies the hash query transaction
// @Accept      json
// @Produce     json
// @Param       hash path     string true "Transaction hash"
// @Success     200  {object} model.Transaction
// @Failure     400  {object} service.ErrRes
// @Router      /transaction/{hash} [get]
func getTransaction(c *gin.Context) {
	hash := c.Param("hash")

	data, err := service.GetTransaction(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query transaction receipt
// @Description specifies the transaction hash to query the transaction receipt
// @Accept      json
// @Produce     json
// @Param       hash path     string true "Transaction hash"
// @Success     200  {object} []model.EventLog
// @Failure     400  {object} service.ErrRes
// @Router      /transaction_logs/{hash} [get]
func getTransactionLogs(c *gin.Context) {
	hash := c.Param("hash")

	data, err := service.GetTransactionLogs(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query internal transaction list
// @Description query internal transaction list in reverse order
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.InternalTxsRes
// @Failure     400       {object} service.ErrRes
// @Router      /transaction/internal/page [get]
func pageInternalTransaction(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.GetInternalTransactions(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query internal transaction list
// @Description specifies the hash query internal transaction
// @Produce     json
// @Param       hash path     string true "tx hash"
// @Success     200  {object} []model.InternalTx
// @Failure     400  {object} service.ErrRes
// @Router      /transaction/internal/{hash} [get]
func getInternalTransaction(c *gin.Context) {
	data, err := service.GetInternalTransaction(c.Param("hash"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query erbie transaction list
// @Description query erbie transaction list in reverse order
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Param       number    query    string false "Block number, if empty, query all"
// @Param       epoch     query    string false "Specify the period id"
// @Param       address   query    string false "nft or snft address, if empty, query all"
// @Param       account   query    string false "Account address, if empty, query all"
// @Param       types     query    string false "erbie tx type,supports multiple types,if empty, query all"
// @Success     200       {object} service.ErbiesRes
// @Failure     400       {object} service.ErrRes
// @Router      /transaction/erbie/page [get]
func pageErbieTransaction(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.FetchErbieTxs(page, size, c.Query("number"), c.Query("address"), c.Query("epoch"), c.Query("account"), c.Query("types"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        transaction
// @Summary     query erbie transaction list
// @Description specifies the hash query internal transaction
// @Produce     json
// @Param       hash path     string true "tx hash"
// @Success     200  {object} model.Erbie
// @Failure     400  {object} service.ErrRes
// @Router      /transaction/erbie/{hash} [get]
func getErbieTransaction(c *gin.Context) {
	data, err := service.GetErbieTransaction(c.Param("hash"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
