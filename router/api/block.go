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
	e.GET("/totals", totals)
	e.GET("/validators", validators)
	e.GET("/locations", locations)
	e.GET("/last_msg", lastMsg)
}

// @Tags        block
// @Summary     query block list
// @Description Query the block list in reverse order of height
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.BlocksRes
// @Failure     400       {object} service.ErrRes
// @Router      /block/page [get]
func pageBlock(c *gin.Context) {
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

	data, err := service.FetchBlocks(page, size)
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
// @Summary     query some total data
// @Description Query the total number of blocks, transactions, accounts, etc.
// @Accept      json
// @Produce     json
// @Success     200 {object} service.Stats
// @Failure     400 {object} service.ErrRes
// @Router      /totals [get]
func totals(c *gin.Context) {
	res, err := service.FetchTotals()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        block
// @Summary     query validator list
// @Description Query validator's information
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} []model.Validator
// @Failure     400       {object} service.ErrRes
// @Router      /validators [get]
func validators(c *gin.Context) {
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
	res, err := service.FetchValidator(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        block
// @Summary     query validator locations
// @Description Query validator's locations
// @Accept      json
// @Produce     json
// @Success     200 {object} []model.Location
// @Failure     400 {object} service.ErrRes
// @Router      /locations [get]
func locations(c *gin.Context) {
	res, err := service.FetchLocations()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        block
// @Summary     query validator msg list
// @Description Query validator's last msg list
// @Accept      json
// @Produce     json
// @Success     200 {object} []service.Msg
// @Router      /last_msg [get]
func lastMsg(c *gin.Context) {
	c.JSON(http.StatusOK, service.GetLastMsg())
}
