package NFT

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/database"
)

func Routers(e *gin.Engine) {
	e.GET("/nft/page", pageNFT)
	e.GET("/nft/tx/page", pageNFTTx)
}

// @Tags  NFT
// @Summary 查询NFT列表
// @Description 按创建时间逆序查询NFT列表
// @Accept json
// @Produce json
// @Param exchanger query string false "交易所，空则查询所有交易所"
// @Param owner query string false "所有者,空则查询所有"
// @Param page query string false "页,默认1"
// @Param page_size query string false "页大小,默认10"
// @Success 200 {object} PageRes
// @Failure 400 {object} ErrRes
// @Router /nft/page [get]
func pageNFT(c *gin.Context) {
	req := struct {
		Page      uint64 `form:"page"`
		PageSize  uint64 `form:"page_size"`
		Exchanger string `form:"exchanger"`
		Owner     string `form:"owner"`
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

	data, count, err := database.FetchUserNFTs(req.Exchanger, req.Owner, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, PageRes{Total: count, NFTs: data})

}

// @Tags  NFT
// @Summary 查询NFT交易列表
// @Description 按创建时间逆序查询NFT交易列表
// @Accept json
// @Produce json
// @Param exchanger query string false "交易所，空则查询所有交易所"
// @Param account query string false "指定帐户,空则查询所有帐户的"
// @Param page query string false "页,默认1"
// @Param page_size query string false "页大小,默认10"
// @Success 200 {object} PageTxRes
// @Failure 400 {object} ErrRes
// @Router /nft/tx/page [get]
func pageNFTTx(c *gin.Context) {
	req := struct {
		Page      uint64 `form:"page"`
		PageSize  uint64 `form:"page_size"`
		Exchanger string `form:"exchanger"`
		Account   string `form:"account"`
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

	data, count, err := database.FetchNFTTxs(req.Exchanger,req.Account, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, PageTxRes{Total: count, NFTTxs: data})

}
