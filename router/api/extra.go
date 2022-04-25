package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/extra"
	"server/service"
)

func Extra(e *gin.Engine) {
	e.GET("/erb_price", erbPrice)
	e.GET("/erb_faucet", erbFaucet)
	e.GET("/exchanger_auth", exchangerAuth)
}

type price struct {
	USD float64 `json:"USD"` //一个ERB美元价格
	CNY float64 `json:"CNY"` //一个ERB人民币价格
}

// @Tags         其他接口
// @Summary      查询ERB价格
// @Description  查询一个ERB价格，1ERB=10^18wei，未能实现ERB价格定义，固定为1ERB=1USD
// @Accept       json
// @Produce      json
// @Success      200  {object}  price
// @Router       /erb_price [get]
func erbPrice(c *gin.Context) {
	c.JSON(http.StatusOK, price{CNY: 6.3, USD: 1})
}

// @Tags         其他接口
// @Summary      请求ERB测试币
// @Description  请求ERB测试币
// @Accept       json
// @Produce      json
// @Param        addr  query     string  true  "地址"
// @Success      200
// @Failure      400  {object}  service.ErrRes
// @Router       /erb_faucet [get]
func erbFaucet(c *gin.Context) {
	addr, err := utils.ParseAddress(c.Query("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	err = extra.SendErb(string(addr), context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

type AuthRes struct {
	Status           uint64 `json:"status"` //2 交易所付费状态正常  其他数字为欠费或者没交费
	ExchangerFlag    bool   `json:"exchanger_flag"`
	ExchangerBalance string `json:"exchanger_balance"`
}

// @Tags         其他接口
// @Summary      查询交易所状态
// @Description  查询交易所状态
// @Accept       json
// @Produce      json
// @Param        addr  query  string  true  "地址"
// @Success      200   {object}  AuthRes
// @Failure      400   {object}  service.ErrRes
// @Router       /exchanger_auth [get]
func exchangerAuth(c *gin.Context) {
	addr, err := utils.ParseAddress(c.Query("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	status, flag, balance, err := extra.ExchangerAuth(string(addr))
	if err != nil {
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, &AuthRes{
		Status:           status,
		ExchangerFlag:    flag,
		ExchangerBalance: balance,
	})
}
