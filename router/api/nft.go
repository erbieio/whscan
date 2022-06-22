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
	e.GET("/snft/page", pageSNFT)
	e.GET("/snft/block", blockSNFT)
	e.GET("/snft_meta/page", pageSNFTAndMeta)
	e.GET("/snft/collection/page", pageSNFTGroup)
	e.GET("/snft/group/:id", groupSNFTs)
}

// @Tags         NFT
// @Summary      查询NFT列表
// @Description  按创建时间逆序查询NFT列表
// @Accept       json
// @Produce      json
// @Param        exchanger      query     string  false  "交易所，空则查询所有交易所"
// @Param        owner          query     string  false  "所有者,空则查询所有"
// @Param        collection_id  query     string  false  "合集id,空则查询所有"
// @Param        page           query     string  false  "页,默认1"
// @Param        page_size      query     string  false  "页大小,默认10"
// @Success      200            {object}  service.UserNFTsRes
// @Failure      400            {object}  service.ErrRes
// @Router       /nft/page [get]
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

	res, err := service.FetchUserNFTs(req.Exchanger, req.CollectionId, strings.ToLower(req.Owner), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags         NFT
// @Summary      查询包含元信息NFT列表
// @Description  按创建时间逆序查询包含元信息NFT列表
// @Accept       json
// @Produce      json
// @Param        exchanger      query     string  false  "交易所，空则查询所有交易所"
// @Param        owner          query     string  false  "所有者,空则查询所有"
// @Param        collection_id  query     string  false  "合集id,空则查询所有"
// @Param        page           query     string  false  "页,默认1"
// @Param        page_size      query     string  false  "页大小,默认10"
// @Success      200            {object}  service.UserNFTsRes
// @Failure      400            {object}  service.ErrRes
// @Router       /nft_meta/page [get]
func pageNFTAndMeta(c *gin.Context) {
	pageNFT(c)
}

// @Tags         NFT
// @Summary      查询NFT交易列表
// @Description  按创建时间逆序查询NFT交易列表
// @Accept       json
// @Produce      json
// @Param        address    query     string  false  "指定NFT地址,空则查询所有地址的"
// @Param        exchanger  query     string  false  "交易所，空则查询所有交易所"
// @Param        account    query     string  false  "指定账户,空则查询所有账户的"
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.NFTTxsRes
// @Failure      400  {object}  service.ErrRes
// @Router       /nft/tx/page [get]
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

// @Tags         NFT
// @Summary      查询SNFT列表
// @Description  按创建时间逆序查询SNFT列表
// @Accept       json
// @Produce      json
// @Param        owner          query     string  false  "所有者,空则查询所有"
// @Param        page           query     string  false  "页,默认1"
// @Param        page_size      query     string  false  "页大小,默认10"
// @Success      200        {object}  service.SNFTsRes
// @Failure      400            {object}  service.ErrRes
// @Router       /snft/page [get]
func pageSNFT(c *gin.Context) {
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

	res, err := service.FetchSNFTs(strings.ToLower(req.Owner), page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags         NFT
// @Summary      查询有元信息SNFT列表
// @Description  按创建时间逆序查询有元信息SNFT列表
// @Accept       json
// @Produce      json
// @Param        collection_id  query     string  false  "合集id,空则查询所有"
// @Param        owner      query     string  false  "所有者,空则查询所有"
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200            {object}  service.SNFTsAndMetaRes
// @Failure      400        {object}  service.ErrRes
// @Router       /snft_meta/page [get]
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

// @Tags         NFT
// @Summary      查询区块SNFT列表
// @Description  查询指定区块的SNFT奖励列表
// @Accept       json
// @Produce      json
// @Param        number  query     string  true  "区块号"
// @Success      200     {object}  []model.SNFT
// @Failure      400        {object}  service.ErrRes
// @Router       /snft/block [get]
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

// @Tags         NFT
// @Summary      分页查询帐户持有合集列表
// @Description  查询指定帐户的持有（有合集里的一个SNFT）的合集列表（包含16个FullNFT信息）
// @Accept       json
// @Produce      json
// @Param        owner      query     string  false  "所有者"
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.SNFTGroupsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /snft/collection/page [get]
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

// @Tags         NFT
// @Summary      查询指定FullNFT的256个SNFT列表
// @Description  查询指定ID的FullNFT下256个SNFT的信息
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "FullNFT编号"
// @Success      200  {object}  []model.SNFT
// @Failure      400     {object}  service.ErrRes
// @Router       /snft/group/{id} [get]
func groupSNFTs(c *gin.Context) {
	id := c.Param("id")
	data, err := service.FullSNFTs(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
