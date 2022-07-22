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
// @Param       limit     query    string false "Limit, time range"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingSNFTRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/snft [get]
func rankingSNFT(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Limit    string `form:"limit"`
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

	res, err := service.RankingSNFT(req.Limit, page, size)
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
// @Param       limit     query    string false "Limit, time range"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingNFTRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/nft [get]
func rankingNFT(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Limit    string `form:"limit"`
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

	res, err := service.RankingNFT(req.Limit, page, size)
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
// @Param       limit     query    string false "Limit, time range"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RankingExchangerRes
// @Failure     400       {object} service.ErrRes
// @Router      /ranking/exchanger [get]
func rankingExchanger(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Limit    string `form:"limit"`
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

	res, err := service.RankingExchanger(req.Limit, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
