package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Epoch 合集API
func Epoch(e *gin.Engine) {
	e.GET("/epoch", pageEpoch)
	e.GET("/epoch/:id", getEpoch)
}

// @Tags         系统NFT期
// @Summary      查询系统NFT期列表
// @Description  按创建时间逆序查询系统NFT期列表
// @Accept       json
// @Produce      json
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.EpochsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /epoch [get]
func pageEpoch(c *gin.Context) {
	req := struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
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

	res, err := service.FetchEpochs(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Tags         系统NFT期
// @Summary      查询系统NFT期
// @Description  指定ID查询系统NFT期信息,包含16个合集信息
// @Accept       json
// @Produce      json
// @Param        id   path      string  false  "期ID，current表示查询当前的期"
// @Success      200  {object}  model.Epoch
// @Failure      400  {object}  service.ErrRes
// @Router       /epoch/{id} [get]
func getEpoch(c *gin.Context) {
	id := c.Param("id")

	data, err := service.GetEpoch(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
