package ERB

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers(e *gin.Engine) {
	e.GET("/erb_price", erbPrice)
}

type Price struct {
	USD float64 `json:"USD"` //一个ERB美元价格
	CNY float64 `json:"CNY"` //一个ERB人民币价格
}

// @Tags         ERB
// @Summary      查询ERB价格
// @Description  查询一个ERB价格，1ERB=10^18wei，未能实现ERB价格定义，固定为1ERB=1USD
// @Accept       json
// @Produce      json
// @Success      200  {object}  Price
// @Router       /erb_price [get]
func erbPrice(c *gin.Context) {
	c.JSON(http.StatusOK, Price{CNY: 6.3, USD: 1})
}
