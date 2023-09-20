package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Block blockAPI
func Block(e *gin.Engine) {
	e.GET("/block/page", pageBlock)
	e.GET("/block/:number", getBlock)
	e.GET("/stats", stats)
	e.GET("/pledge/page", pagePledge)
}

// @Tags        block
// @Summary     query block list
// @Description Query the block list in reverse order of height
// @Accept      json
// @Produce     json
// @Param       filter    query    string false "filter block, 1: black hole; 2: penalty; other: all"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.BlocksRes
// @Failure     400       {object} service.ErrRes
// @Router      /block/page [get]
func pageBlock(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.FetchBlocks(page, size, c.Query("filter"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        block
// @Summary     query block
// @Description specifies the number query block
// @Accept      json
// @Produce     json
// @Param       number path     string true "Block number"
// @Success     200    {object} model.Block
// @Failure     400    {object} service.ErrRes
// @Router      /block/{number} [get]
func getBlock(c *gin.Context) {
	number := c.Param("number")

	data, err := service.GetBlock(number)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        block
// @Summary     query some stats data
// @Description Query the total number of blocks, transactions, accounts, etc.
// @Accept      json
// @Produce     json
// @Success     200 {object} model.Stats
// @Failure     400 {object} service.ErrRes
// @Router      /stats [get]
func stats(c *gin.Context) {
	c.JSON(http.StatusOK, service.GetStats())
}

// @Tags        block
// @Summary     query pledge list
// @Description Query pledge list
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Param       staker    query    string false "staker address"
// @Param       validator query    string false "validator address"
// @Success     200       {object} service.PledgesRes
// @Failure     400       {object} service.ErrRes
// @Router      /pledge/page [get]
func pagePledge(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	data, err := service.FetchPledges(c.Query("staker"), c.Query("validator"), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
