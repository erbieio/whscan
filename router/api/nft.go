package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

func NFT(e *gin.Engine) {
	e.GET("/nft/page", pageNFT)
	e.GET("/nft/:addr", getNFT)
	e.GET("/snft/page", pageSNFT)
	e.GET("/snft/:addr", getSNFT)
	e.GET("/snft/block", blockSNFT)
	e.GET("/snft_meta/page", pageSNFTAndMeta)
}

// @Tags        NFT
// @Summary     query NFT list
// @Description Query the NFT list in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       owner     query    string false "Owner, if empty, query all"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.NFTsRes
// @Failure     400       {object} service.ErrRes
// @Router      /nft/page [get]
func pageNFT(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchNFTs(strings.ToLower(c.Query("owner")), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     query SNFT list
// @Description Query the SNFT list in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       sort      query    string false "sort, 1:level priority,none:default"
// @Param       owner     query    string false "Owner, if empty, query all"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.SNFTsRes
// @Failure     400       {object} service.ErrRes
// @Router      /snft/page [get]
func pageSNFT(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchSNFTs(c.Query("sort"), c.Query("owner"), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     query one NFT
// @Description Query one NFT by address
// @Accept      json
// @Produce     json
// @Param       addr path     string true "Address"
// @Success     200  {object} model.NFT
// @Failure     400  {object} service.ErrRes
// @Router      /nft/{addr} [get]
func getNFT(c *gin.Context) {
	res, err := service.GetNFT(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     query one SNFT
// @Description Query one SNFT by address
// @Accept      json
// @Produce     json
// @Param       addr path     string true "Address"
// @Success     200  {object} service.SNFTRes
// @Failure     400  {object} service.ErrRes
// @Router      /snft/{addr} [get]
func getSNFT(c *gin.Context) {
	res, err := service.GetSNFT(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     Query a list of SNFTs with meta information
// @Description Query the list of SNFTs with meta information in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       owner     query    string false "Owner, if empty, query all"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.SNFTsAndMetaRes
// @Failure     400       {object} service.ErrRes
// @Router      /snft_meta/page [get]
func pageSNFTAndMeta(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	res, err := service.FetchSNFTsAndMeta(strings.ToLower(c.Query("owner")), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     Query blocks SNFT list
// @Description Query the list of SNFT rewards for the specified block
// @Accept      json
// @Produce     json
// @Param       number query    string true "Block number"
// @Success     200    {object} []model.SNFT
// @Failure     400    {object} service.ErrRes
// @Router      /snft/block [get]
func blockSNFT(c *gin.Context) {
	data, err := service.BlockSNFTs(c.Query("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
