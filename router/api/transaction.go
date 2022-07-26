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
// @Success     200       {object} service.TransactionsRes
// @Failure     400       {object} service.ErrRes
// @Router      /transaction/page [get]
func pageTransaction(c *gin.Context) {
	req := struct {
		Page     *int    `form:"page"`
		PageSize *int    `form:"page_size"`
		Number   *string `form:"number"`
		Addr     *string `form:"addr"`
	}{}
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	page, size, err := utils.ParsePage(req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	data, err := service.FetchTransactions(page, size, req.Number, req.Addr)
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
// @Success     200  {object} service.TransactionRes
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
// @Success     200  {object} []model.Log
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
