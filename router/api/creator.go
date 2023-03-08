package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Creator creatorAPI
func Creator(e *gin.Engine) {
	e.GET("/creator", pageCreator)
	e.GET("/creator/:addr", getCreator)
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
// @Router      /creator [get]
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
