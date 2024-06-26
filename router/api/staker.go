package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

func Staker(e *gin.Engine) {
	e.GET("/staker/page", pageStaker)
	e.GET("/staker/:addr", getStaker)
}

// @Tags        Staker
// @Summary     Query the list of stakers
// @Description Query the list of stakers in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.StakersRes
// @Failure     400       {object} service.ErrRes
// @Router      /staker/page [get]
func pageStaker(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchStakers(c.Query("order"), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        Staker
// @Summary     query staker
// @Description Query staker by address
// @Accept      json
// @Produce     json
// @Param       addr path     string true "staker address"
// @Success     200  {object} model.Staker
// @Failure     400  {object} service.ErrRes
// @Router      /staker/{addr} [get]
func getStaker(c *gin.Context) {
	address := c.Param("addr")
	if address == "" {
		address = c.Query("addr")
	}
	addr, err := utils.ParseAddress([]byte(address))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	data, err := service.FindStaker(addr)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
