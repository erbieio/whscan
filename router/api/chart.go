package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/service"
)

// Chart ChartAPI
func Chart(e *gin.Engine) {
	e.GET("/chart/line", lineChart)
}

// @Tags        chart
// @Summary     query charts
// @Description query charts
// @Accept      json
// @Produce     json
// @Param       limit query    string false "Limit, default 10"
// @Success     200   {object} service.LineChartRes
// @Failure     400   {object} service.ErrRes
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
