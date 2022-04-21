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
}

// @Tags         交易所
// @Summary      查询交易所列表
// @Description  按创建时间逆序查询交易所列表
// @Accept       json
// @Produce      json
// @Param        name       query     string  false  "交易所名称,空则查询所有交易所"
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.CollectionsRes
// @Failure      400        {object}  service.ErrRes
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

// @Tags         交易所
// @Summary      查询交易所
// @Description  按地址查询交易所
// @Accept       json
// @Produce      json
// @Param        addr  path      string  true  "交易所地址"
// @Success      200   {object}  model.Exchanger
// @Failure      400   {object}  service.ErrRes
// @Router       /exchanger/{addr} [get]
func getExchanger(c *gin.Context) {
	address := c.Param("addr")
	if address == "" {
		address = c.Query("addr")
	}
	if len(address) != 42 {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: "地址不是42个字符"})
		return
	}
	data, err := service.FindExchanger(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
