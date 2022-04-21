package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Block 区块API
func Block(e *gin.Engine) {
	e.GET("/block/page", pageBlock)
	e.GET("/block/:number", getBlock)
}

// @Tags         区块
// @Summary      查询区块列表
// @Description  按高度逆序查询区块列表
// @Accept       json
// @Produce      json
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.BlocksRes
// @Failure      400        {object}  service.ErrRes
// @Router       /block/page [get]
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

// @Tags         区块
// @Summary      查询区块
// @Description  指定number查询区块
// @Accept       json
// @Produce      json
// @Param        number  path      string  true  "区块号"
// @Success      200     {object}  model.Block
// @Failure      400     {object}  service.ErrRes
// @Router       /block/{number} [get]
func getBlock(c *gin.Context) {
	number := c.Param("number")

	data, err := service.GetBlock(number)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
