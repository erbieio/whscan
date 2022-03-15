package extra

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/ethhelper"
)

func Routers(e *gin.Engine) {
	e.GET("/extra/checkAuth", checkAuth)
}

// @Tags  其他接口
// @Summary 查询交易所状态
// @Description 查询交易所状态
// @Accept json
// @Produce json
// @Param body body CheckAuthReq true "body"
// @Success 200 {object} CheckAuthRes
// @Failure 400 {object} ErrRes
// @Router /extra/checkAuth [get]
func checkAuth(c *gin.Context) {
	var req CheckAuthReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}
	res, err := ethhelper.CheckAuth(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, CheckAuthRes{Code: 0, Msg: "ok", Result: res})
}
