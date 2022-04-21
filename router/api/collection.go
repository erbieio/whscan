package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Collection 合集API
func Collection(e *gin.Engine) {
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
// @Success      200        {object}  service.CollectionsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /collection/page [get]
func pageCollection(c *gin.Context) {
	req := struct {
		Page      *int   `form:"page"`
		PageSize  *int   `form:"page_size"`
		Exchanger string `form:"exchanger"`
		Creator   string `form:"creator"`
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

	res, err := service.FetchCollections(req.Exchanger, req.Creator, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags         NFT合集
// @Summary      查询NFT合集
// @Description  指定ID查询NFT合集信息
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "合集ID"
// @Success      200  {object}  model.Collection
// @Failure      400  {object}  service.ErrRes
// @Router       /collection/{id} [get]
func getCollection(c *gin.Context) {
	id := c.Param("id")

	data, err := service.GetCollection(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
