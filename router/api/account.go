package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Account accountAPI
func Account(e *gin.Engine) {
	e.GET("/account/page", pageAccount)
	e.GET("/account/:addr", getAccount)
	e.GET("/account/growth_rate", getAccountGrowthRate)
	e.GET("/account/total_num", getAccountTotalNum)
}

// @Tags        account
// @Summary     query top accounts
// @Description set the account ranking according to the amount of coins held
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.AccountsRes
// @Failure     400       {object} service.ErrRes
// @Router      /account/page [get]
func pageAccount(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchAccounts(page, size, c.Query("order"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        account
// @Summary     query one account
// @Description Query the account information of the specified address
// @Accept      json
// @Produce     json
// @Param       addr path     string true "account address"
// @Success     200  {object} service.AccountRes
// @Failure     400  {object} service.ErrRes
// @Router      /account/{addr} [get]
func getAccount(c *gin.Context) {
	res, err := service.GetAccount(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        account growth rate
// @Summary     query account growth rate
// @Description query account growth rate
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} float64
// @Failure     400  {object} service.ErrRes
// @Router      /account/growth_rate [get]
func getAccountGrowthRate(c *gin.Context) {
	res, err := service.GetAccountGrowthRate()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        account total number
// @Summary     query account growth rate
// @Description query account total number
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /account/total_num [get]
func getAccountTotalNum(c *gin.Context) {
	res, err := service.GetAccountTotalNum()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
