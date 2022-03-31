package exchanger

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/database"
)

func Routers(e *gin.Engine) {
	e.GET("/exchanger/page", pageExchanger)
	e.GET("/exchanger/get", getExchanger)
}

// @Tags  交易所
// @Summary 查询交易所列表
// @Description 按创建时间逆序查询交易所列表
// @Accept json
// @Produce json
// @Param name query string false "交易所名称,空则查询所有交易所"
// @Param page query string false "页,默认1"
// @Param page_size query string false "页大小,默认10"
// @Success 200 {object} PageRes
// @Failure 400 {object} ErrRes
// @Router /exchanger/page [get]
func pageExchanger(c *gin.Context) {
	req := struct {
		Page     uint64 `form:"page"`
		PageSize uint64 `form:"page_size"`
		Name     string `form:"name"`
	}{}
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	data, count, err := database.FetchExchangers(req.Name, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	var total int64
	if req.Name == "" {
		total, err = database.YesterdayExchangerTotal()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, PageRes{Total: count, YesterdayTotal: total, Exchangers: data})

}

// @Tags  交易所
// @Summary 查询交易所
// @Description 按地址查询交易所
// @Accept json
// @Produce json
// @Param addr query string true "交易所地址"
// @Success 200 {object} database.Exchanger
// @Failure 400 {object} ErrRes
// @Router /exchanger/get [get]
func getExchanger(c *gin.Context) {
	address := c.Query("addr")
	if len(address) != 42 {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: "地址不是42个字符"})
		return
	}
	data, err := database.FindExchanger(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
