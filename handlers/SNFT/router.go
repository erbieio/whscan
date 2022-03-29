package SNFT

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/database"
)

func Routers(e *gin.Engine) {
	e.GET("/snft/page", pageSNFT)
}

// @Tags  SNFT
// @Summary 查询SNFT列表
// @Description 按创建时间逆序查询SNFT列表
// @Accept json
// @Produce json
// @Param owner query string false "所有者,空则查询所有"
// @Param page query string false "页,默认1"
// @Param page_size query string false "页大小,默认10"
// @Success 200 {object} PageRes
// @Failure 400 {object} ErrRes
// @Router /snft/page [get]
func pageSNFT(c *gin.Context) {
	req := struct {
		Page     uint64 `form:"page"`
		PageSize uint64 `form:"page_size"`
		Owner    string `form:"owner"`
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

	data, count, err := database.FetchOfficialNFTs(req.Owner, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, PageRes{Total: count, NFTs: data})

}
