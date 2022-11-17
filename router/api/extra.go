package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

func Extra(e *gin.Engine) {
	e.GET("/erb_price", erbPrice)
	e.GET("/exchanger_auth", exchangerAuth)
	e.POST("/subscription", subscribe)
	e.GET("/subscription", pageSubscribe)
}

type price struct {
	USD float64 `json:"USD"` //The price of an ERB in USD
	CNY float64 `json:"CNY"` //The price of an ERB in RMB
}

// @Tags        other interfaces
// @Summary     query ERB price
// @Description Query an ERB price, 1ERB=10^18wei, failed to implement the ERB price definition, fixed at 1ERB=0.5USD
// @Accept      json
// @Produce     json
// @Success     200 {object} price
// @Router      /erb_price [get]
func erbPrice(c *gin.Context) {
	c.JSON(http.StatusOK, price{CNY: 3.2, USD: 0.5})
}

type AuthRes struct {
	Status           uint64 `json:"status"` //2 The payment status of the exchange is normal, other numbers are arrears or no payment
	ExchangerFlag    bool   `json:"exchanger_flag"`
	ExchangerBalance string `json:"exchanger_balance"`
}

// @Tags        other interfaces
// @Summary     query exchange status
// @Description query exchange status
// @Accept      json
// @Produce     json
// @Param       addr query    string true "address"
// @Success     200  {object} AuthRes
// @Failure     400  {object} service.ErrRes
// @Router      /exchanger_auth [get]
func exchangerAuth(c *gin.Context) {
	addr, err := utils.ParseAddress([]byte(c.Query("addr")))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	status, flag, balance, err := service.ExchangerAuth(string(addr))
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

type Subscribe struct {
	Email string `form:"email"`
}

// @Tags        other interfaces
// @Summary     subscribe email
// @Description Enter the email address to receive the latest event notifications
// @Accept      json
// @Produce     json
// @Param       _ body Subscribe true "Mailbox"
// @Success     200
// @Failure     400 {object} service.ErrRes
// @Router      /subscription [post]
func subscribe(c *gin.Context) {
	req := Subscribe{}
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	if !utils.VerifyEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: "The email format is incorrect"})
		return
	}

	err = service.SaveSubscription(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// @Tags        other interfaces
// @Summary     Query the list of subscription mailboxes
// @Description Query the list of subscription mailboxes
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} []model.Subscription
// @Failure     400       {object} service.ErrRes
// @Router      /subscription [get]
func pageSubscribe(c *gin.Context) {
	req := struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
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

	res, err := service.FetchSubscriptions(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
