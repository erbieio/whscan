package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Transaction 交易API
func Transaction(e *gin.Engine) {
	e.GET("/transaction/page", pageTransaction)
	e.GET("/transaction/:hash", getTransaction)
	e.GET("/transaction_logs/:hash", getTransactionLogs)
}

// @Tags         交易
// @Summary      查询交易列表
// @Description  逆序查询交易列表
// @Accept       json
// @Produce      json
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Param        number     query     string  false  "区块号，空则查询所有"
// @Param        addr       query     string  false  "帐户地址，空则查询所有"
// @Success      200        {object}  service.TransactionsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /transaction/page [get]
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

// @Tags         交易
// @Summary      查询交易
// @Description  指定hash查询交易
// @Accept       json
// @Produce      json
// @Param        hash  path      string  true  "交易哈希"
// @Success      200   {object}  model.Transaction
// @Failure      400   {object}  service.ErrRes
// @Router       /transaction/{hash} [get]
func getTransaction(c *gin.Context) {
	hash := c.Param("hash")

	data, err := service.GetTransaction(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags         交易
// @Summary      查询交易收据
// @Description  指定交易hash查询交易收据
// @Accept       json
// @Produce      json
// @Param        hash  path      string  true  "交易哈希"
// @Success      200   {object}  []model.Log
// @Failure      400   {object}  service.ErrRes
// @Router       /transaction_logs/{hash} [get]
func getTransactionLogs(c *gin.Context) {
	hash := c.Param("hash")

	data, err := service.GetTransactionLogs(hash)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
