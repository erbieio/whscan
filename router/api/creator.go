package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Creator creatorAPI
func Creator(e *gin.Engine) {
	e.GET("/creator/page", pageCreator)
	e.GET("/creator/:addr", getCreator)
	e.GET("/creator/top", topCreator)
}

// @Tags        creator
// @Summary     Query the creator list
// @Description Query the creator list, page
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Success     200       {object} service.CreatorsRes
// @Failure     400       {object} service.ErrRes
// @Router      /creator/page [get]
func pageCreator(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Order    string `form:"order"`
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

	res, err := service.FetchCreators(page, size, req.Order)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        creator
// @Summary     Query a creator
// @Description specifies the address to query the creator
// @Accept      json
// @Produce     json
// @Param       addr path     string false "creator address"
// @Success     200  {object} model.Creator
// @Failure     400  {object} service.ErrRes
// @Router      /creator/{addr} [get]
func getCreator(c *gin.Context) {
	data, err := service.GetCreator(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        creator
// @Summary     Query the top creators
// @Description Query the top creators
// @Accept      json
// @Produce     json
// @Param       size query    integer false "request number, default 10"
// @Success     200  {array}  model.Creator
// @Failure     400  {object} service.ErrRes
// @Router      /creator/top [get]
func topCreator(c *gin.Context) {
	req := struct {
		Page int `form:"size"`
	}{}
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	data, err := service.TopCreators(req.Page)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
