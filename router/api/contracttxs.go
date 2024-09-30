package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/common/utils"
	"server/service"
)

// Contract ContractAPI
func ContractTxs(e *gin.Engine) {
	e.GET("/contract_tx/page", pageContractTxs)
	e.GET("/contract_tx/total_num", getContractTxTotalNum)
}

// @Tags        contract txs
// @Summary     query top contract txs
// @Description set the contract ranking according to create time
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.ContracttxsRes
// @Failure     400       {object} service.ErrRes
// @Router      /contract_tx/page [get]
func pageContractTxs(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	addr := c.Query("addr")

	res, err := service.FetchContractTxs(addr, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        contract txs total number
// @Summary     query contract txs total number
// @Description query contract txs total number
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /contract_tx/total_num [get]
func getContractTxTotalNum(c *gin.Context) {
	res, err := service.GetContractTxTotalNum()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
