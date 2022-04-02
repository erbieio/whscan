package collection

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/database"
)

func Routers(e *gin.Engine) {
	e.GET("/collection/page", pageCollection)
	e.GET("/collection/:id", getCollection)
}

// @Tags         NFT合集
// @Summary      查询NFT合集列表
// @Description  按创建区块高度逆序查询NFT合集列表
// @Accept       json
// @Produce      json
// @Param        exchanger  query     string  false  "交易所，空则查询所有交易所"
// @Param        creator    query     string  false  "创建者,空则查询所有"
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  PageRes
// @Failure      400        {object}  ErrRes
// @Router       /collection/page [get]
func pageCollection(c *gin.Context) {
	req := struct {
		Page      uint64 `form:"page"`
		PageSize  uint64 `form:"page_size"`
		Exchanger string `form:"exchanger"`
		Creator   string `form:"creator"`
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

	data, count, err := database.FetchCollections(req.Exchanger, req.Creator, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, PageRes{Total: count, Collections: data})
}

// @Tags         NFT合集
// @Summary      查询NFT合集
// @Description  指定ID查询NFT合集信息
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "合集ID"
// @Success      200  {object}  database.Collection
// @Failure      400  {object}  ErrRes
// @Router       /collection/{id} [get]
func getCollection(c *gin.Context) {
	id := c.Param("id")

	data, err := database.GetCollection(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
