package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Collection CollectionAPI
func Collection(e *gin.Engine) {
	e.GET("/collection/page", pageCollection)
	e.GET("/collection/:id", getCollection)
}

// @Tags         NFT Collection
// @Summary      Query the list of NFT collections
// @Description  Query NFT collection list in reverse order of created block height
// @Accept       json
// @Produce      json
// @Param        exchanger  query     string  false  "Exchange, if empty, query all exchanges"
// @Param        creator    query     string  false  "Creator, if empty, query all"
// @Param        type       query     string  false  "Type: all, nft, snft, default all"
// @Param        page       query     string  false  "Page, default 1"
// @Param        page_size  query     string  false  "Page size, default 10"
// @Success      200        {object}  service.CollectionsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /collection/page [get]
func pageCollection(c *gin.Context) {
	req := struct {
		Page      *int   `form:"page"`
		PageSize  *int   `form:"page_size"`
		Exchanger string `form:"exchanger"`
		Creator   string `form:"creator"`
		Type      string `form:"type"`
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

	res, err := service.FetchCollections(req.Exchanger, req.Creator, req.Type, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags         NFT Collection
// @Summary      query NFT collection
// @Description  specifies the ID to query the NFT collection information
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Collection ID"
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
