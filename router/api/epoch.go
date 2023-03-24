package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Epoch epochAPI
func Epoch(e *gin.Engine) {
	e.GET("/epoch", pageEpoch)
	e.GET("/epoch/:id", getEpoch)
}

// @Tags        system NFT period
// @Summary     Query the system NFT period list
// @Description Query the system NFT period list in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       creator   query    string false "only return the creator's epoch"
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.EpochsRes
// @Failure     400       {object} service.ErrRes
// @Router      /epoch [get]
func pageEpoch(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchEpochs(c.Query("creator"), c.Query("order"), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        system NFT period
// @Summary     Query system NFT period
// @Description specifies the ID to query the NFT period information of the system, including 16 collection information
// @Accept      json
// @Produce     json
// @Param       id  path     string false "Period ID, current means query the current period"
// @Success     200 {object} model.Epoch
// @Failure     400 {object} service.ErrRes
// @Router      /epoch/{id} [get]
func getEpoch(c *gin.Context) {
	id := c.Param("id")

	data, err := service.GetEpoch(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
