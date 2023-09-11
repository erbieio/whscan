package api

import (
	"net/http"
	"server/common/utils"

	"github.com/gin-gonic/gin"
	"server/service"
)

func Extra(e *gin.Engine) {
	e.GET("/exec_sql", execSql)
	e.GET("/erb_price", erbPrice)
	e.GET("/slashings", pageSlashing)
}

// @Tags        extra
// @Summary     exec sql statement
// @Description execute sql statement and return the result, only read
// @Accept      json
// @Produce     json
// @Param       key query    string false "admin key"
// @Param       sql query    string false "sql statement"
// @Success     200 {array}  string
// @Failure     400 {object} service.ErrRes
// @Router      /exec_sql [get]
func execSql(c *gin.Context) {
	if c.Query("key") != "123456789kd.wl" {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: "key error, not allow"})
		return
	}
	res, err := service.ExecSql(c.Query("sql"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

type price struct {
	USD float64 `json:"USD"` //The price of an ERB in USD
	CNY float64 `json:"CNY"` //The price of an ERB in RMB
}

// @Tags        extra
// @Summary     query ERB price
// @Description Query an ERB price, 1ERB=10^18wei, failed to implement the ERB price definition, fixed at 1ERB=0.5USD
// @Accept      json
// @Produce     json
// @Success     200 {object} price
// @Router      /erb_price [get]
func erbPrice(c *gin.Context) {
	c.JSON(http.StatusOK, price{CNY: 3.2, USD: 0.5})
}

// @Tags        extra
// @Summary     query slashing list
// @Description Query slashing list
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Param       address   query    string false "penalty address"
// @Param       number    query    string false "penalty block number"
// @Param       reason    query    string false "penalty reason, 1: no block; 2: multi-signature; address: validator penalty"
// @Success     200       {object} service.SlashingsRes
// @Failure     400       {object} service.ErrRes
// @Router      /slashings [get]
func pageSlashing(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.FetchSlashings(c.Query("address"), c.Query("number"), c.Query("reason"), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
