package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/service"
)

// Chart ChartAPI
func Chart(e *gin.Engine) {
	e.GET("/chart/line", lineChart)
	e.GET("/chart/tx", txChart)
	e.GET("/chart/nft", nftChart)
}

// @Tags        chart
// @Summary     query charts
// @Description query charts
// @Accept      json
// @Produce     json
// @Param       limit query    string false "Limit, default 10"
// @Success     200   {object} service.LineChartRes
// @Failure     400 {object} service.ErrRes
// @Router      /chart/line [get]
func lineChart(c *gin.Context) {
	req := struct {
		Limit *int `form:"limit"`
	}{}
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	res, err := service.LineChart(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        chart
// @Summary     query 24h tx growth charts
// @Description query 24h tx growth charts
// @Accept      json
// @Produce     json
// @Success     200 {object} service.TxChartRes
// @Failure     400 {object} service.ErrRes
// @Router      /chart/tx [get]
func txChart(c *gin.Context) {
	res, err := service.TxChart()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        chart
// @Summary     query 24h nft growth charts
// @Description query 24h nft growth charts
// @Accept      json
// @Produce     json
// @Success     200 {object} service.NFTChartRes
// @Failure     400   {object} service.ErrRes
// @Router      /chart/nft [get]
func nftChart(c *gin.Context) {
	res, err := service.NFTChart()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
