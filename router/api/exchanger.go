package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

func Exchanger(e *gin.Engine) {
	e.GET("/exchanger/page", pageExchanger)
	e.GET("/exchanger/:addr", getExchanger)
	e.GET("/exchangers", exchangers)
	e.GET("/exchanger/tx_count/:addr", getExchangerTxCount)
}

// @Tags         Exchange
// @Summary      Query the list of exchanges
// @Description  Query the list of exchanges in reverse order of creation time
// @Accept       json
// @Produce      json
// @Param        name       query     string  false  "Exchange name, if empty, query all exchanges"
// @Param        page       query     string  false  "Page, default 1"
// @Param        page_size  query     string  false  "Page size, default 10"
// @Success      200        {object}  service.ExchangersRes
// @Failure      400   {object}  service.ErrRes
// @Router       /exchanger/page [get]
func pageExchanger(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Name     string `form:"name"`
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

	res, err := service.FetchExchangers(req.Name, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags         Exchange
// @Summary      query exchange
// @Description  Query exchanges by address
// @Accept       json
// @Produce      json
// @Param        addr  path      string  true  "Exchange address"
// @Success      200   {object}  service.ExchangerRes
// @Failure      400   {object}  service.ErrRes
// @Router       /exchanger/{addr} [get]
func getExchanger(c *gin.Context) {
	address := c.Param("addr")
	if address == "" {
		address = c.Query("addr")
	}
	addr, err := utils.ParseAddress(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	data, err := service.FindExchanger(addr)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags         Exchange
// @Summary      Query the list of exchanges
// @Description  Query the list of exchanges in reverse order of creation time
// @Accept       json
// @Produce      json
// @Param        page       query     string  false  "Page, default 1"
// @Param        page_size  query     string  false  "Page size, default 10"
// @Success      200        {object}  []service.ExchangerRes
// @Failure      400        {object}  service.ErrRes
// @Router       /exchangers [get]
func exchangers(c *gin.Context) {
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

	res, err := service.Exchangers(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags         Exchange
// @Summary      Query the exchanges tx count
// @Description  Query the exchanges tx count chart
// @Accept       json
// @Produce      json
// @Param        addr  path      string  true  "Exchanger address"
// @Success      200   {object}  []service.ExchangerTxCountRes
// @Failure      400        {object}  service.ErrRes
// @Router       /exchanger/tx_count/{addr} [get]
func getExchangerTxCount(c *gin.Context) {
	address := c.Param("addr")
	addr, err := utils.ParseAddress(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	data, err := service.ExchangerTxCount(addr)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
