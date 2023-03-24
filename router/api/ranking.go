package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

func Ranking(e *gin.Engine) {
	e.GET("/ranking/nft", rankingNFT)
	e.GET("/ranking/snft", rankingSNFT)
	e.GET("/ranking/exchanger", rankingExchanger)
}

// @Tags        Ranking
// @Summary     query SNFT ranking
// @Description SNFT ranking for a specified time range
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingSNFTRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/snft [get]
func rankingSNFT(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.RankingSNFT(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        Ranking
// @Summary     query NFT ranking
// @Description NFT ranking for a specified time range
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingNFTRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/nft [get]
func rankingNFT(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.RankingNFT(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        Ranking
// @Summary     query exchanger ranking
// @Description Exchanger ranking for a specified time range
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingExchangerRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/exchanger [get]
func rankingExchanger(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.RankingExchanger(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
