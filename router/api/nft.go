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
	e.GET("/nft_meta/page", pageNFTAndMeta)
	e.GET("/nft/tx/page", pageNFTTx)
	e.GET("/nft/tx/:hash", getNFTTx)
	e.GET("/snft/page", pageSNFT)
	e.GET("/snft_com/page", pageComSNFT)
	e.GET("/snft/recycle_tx", getRecycleTx)
	e.GET("/nft/:addr", getNFT)
	e.GET("/snft/:addr", getSNFT)
	e.GET("/snft/block", blockSNFT)
	e.GET("/snft_meta/page", pageSNFTAndMeta)
	e.GET("/snft/collection/page", pageSNFTGroup)
	e.GET("/snft/group/:id", groupSNFTs)
}

// @Tags        NFT
// @Summary     query NFT list
// @Description Query the NFT list in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       exchanger     query    string false "Exchange, if empty, query all exchanges"
// @Param       owner         query    string false "Owner, if empty, query all"
// @Param       collection_id query    string false "collection id, if empty, query all"
// @Param       page          query    string false "Page, default 1"
// @Param       page_size     query    string false "Page size, default 10"
// @Success     200           {object} service.NFTsRes
// @Failure     400           {object} service.ErrRes
// @Router      /nft/page [get]
func pageNFT(c *gin.Context) {
	req := struct {
		Page         *int   `form:"page"`
		PageSize     *int   `form:"page_size"`
		Exchanger    string `form:"exchanger"`
		CollectionId string `form:"collection_id"`
		Owner        string `form:"owner"`
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

	res, err := service.FetchNFTs(req.Exchanger, req.CollectionId, strings.ToLower(req.Owner), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     query contains meta information NFT list
// @Description query the NFT list containing meta information in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       exchanger     query    string false "Exchange, if empty, query all exchanges"
// @Param       owner         query    string false "Owner, if empty, query all"
// @Param       collection_id query    string false "collection id, if empty, query all"
// @Param       page          query    string false "Page, default 1"
// @Param       page_size     query    string false "Page size, default 10"
// @Success     200           {object} service.NFTsRes
// @Failure     400           {object} service.ErrRes
// @Router      /nft_meta/page [get]
func pageNFTAndMeta(c *gin.Context) {
	pageNFT(c)
}

// @Tags        NFT
// @Summary     Query NFT transaction list
// @Description Query the list of NFT transactions in reverse order of creation time
// @Accept      json
// @Produce     json
// @Param       address   query    string false "Specify the NFT address, if empty, query all addresses"
// @Param       exchanger query    string false "Exchange, if empty, query all exchanges"
// @Param       account   query    string false "Specify an account, if empty, query all accounts"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.NFTTxsRes
// @Failure     400       {object} service.ErrRes
// @Router      /nft/tx/page [get]
func pageNFTTx(c *gin.Context) {
	req := struct {
		Page      *int   `form:"page"`
		PageSize  *int   `form:"page_size"`
		Address   string `form:"address"`
		Exchanger string `form:"exchanger"`
		Account   string `form:"account"`
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

	res, err := service.FetchNFTTxs(req.Address, req.Exchanger, req.Account, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     Query NFT transaction
// @Description Query the NFT of SNFT transactions by hash
// @Accept      json
// @Produce     json
// @Param       hash path     string true "transaction hash"
// @Success     200  {object} model.NFTTx
// @Failure     400  {object} service.ErrRes
// @Router      /nft/tx/{hash} [get]
func getNFTTx(c *gin.Context) {
	res, err := service.GetNFTTx(c.Param("hash"))
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
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Owner    string `form:"owner"`
		Sort     int    `form:"sort"`
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

	res, err := service.FetchSNFTs(req.Sort, strings.ToLower(req.Owner), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags        NFT
// @Summary     query Composable SNFT list (new /snft/page)
// @Description Query the Composable SNFT list in reverse order of creation time
// @Deprecated
// @Accept  json
// @Produce json
// @Param   sort      query    string false "sort, 1:level priority,none:default"
// @Param   owner     query    string false "Owner, if empty, query all"
// @Param   page      query    string false "Page, default 1"
// @Param   page_size query    string false "Page size, default 10"
// @Success 200       {object} service.SNFTsRes
// @Failure 400       {object} service.ErrRes
// @Router  /snft_com/page [get]
func pageComSNFT(c *gin.Context) {
	pageSNFT(c)
}

// @Tags        NFT
// @Summary     query recycle transaction
// @Description Query one SNFT recycle transaction by address or tx hash
// @Accept      json
// @Produce     json
// @Param       hash query    string false "recycle tx hash"
// @Param       addr query    string false "recycle address"
// @Success     200  {object} model.NFTTx
// @Failure     400  {object} service.ErrRes
// @Router      /snft/recycle_tx [get]
func getRecycleTx(c *gin.Context) {
	res, err := service.GetRecycleTx(c.Query("hash"), c.Query("addr"))
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
// @Success     200  {object} service.NFTRes
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
// @Param       collection_id query    string false "collection id, if empty, query all"
// @Param       exchanger     query    string false "exchanger, if empty, query all"
// @Param       owner         query    string false "Owner, if empty, query all"
// @Param       page          query    string false "Page, default 1"
// @Param       page_size     query    string false "Page size, default 10"
// @Success     200           {object} service.SNFTsAndMetaRes
// @Failure     400           {object} service.ErrRes
// @Router      /snft_meta/page [get]
func pageSNFTAndMeta(c *gin.Context) {
	req := struct {
		Page         *int   `form:"page"`
		PageSize     *int   `form:"page_size"`
		CollectionId string `form:"collection_id"`
		Owner        string `form:"owner"`
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

	res, err := service.FetchSNFTsAndMeta(strings.ToLower(req.Owner), req.CollectionId, page, size)
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
	req := struct {
		Number uint64 `form:"number"`
	}{}
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	data, err := service.BlockSNFTs(req.Number)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        NFT
// @Summary     paging query account holding collection list
// @Description Query the collection list (including 16 FNFT information) held by the specified account (with one SNFT in the collection)
// @Accept      json
// @Produce     json
// @Param       owner     query    string false "owner"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.SNFTGroupsRes
// @Failure     400       {object} service.ErrRes
// @Router      /snft/collection/page [get]
func pageSNFTGroup(c *gin.Context) {
	req := struct {
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
		Owner    string `form:"owner"`
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
	data, err := service.FindSNFTGroups(strings.ToLower(req.Owner), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags        NFT
// @Summary     Query the list of 16 SNFTs of the specified FNFT
// @Description Query the information of 16 SNFTs under the FNFT of the specified ID
// @Accept      json
// @Produce     json
// @Param       id  path     string true "FNFT ID"
// @Success     200 {object} []model.SNFT
// @Failure     400 {object} service.ErrRes
// @Router      /snft/group/{id} [get]
func groupSNFTs(c *gin.Context) {
	id := c.Param("id")
	data, err := service.FNFTs(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
