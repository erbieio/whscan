package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Reward 奖励API
func Reward(e *gin.Engine) {
	e.GET("/reward", pageReward)
	e.GET("/reward/:block", blockReward)
}

// @Tags         奖励
// @Summary      查询奖励列表
// @Description  逆序查询奖励列表
// @Accept       json
// @Produce      json
// @Param        page       query     string  false  "页,默认1"
// @Param        page_size  query     string  false  "页大小,默认10"
// @Success      200        {object}  service.RewardsRes
// @Failure      400        {object}  service.ErrRes
// @Router       /reward [get]
func pageReward(c *gin.Context) {
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

	data, err := service.FetchRewards(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// @Tags         奖励
// @Summary      查询奖励
// @Description  指定区块查询奖励
// @Accept       json
// @Produce      json
// @Param        block  path      string  true  "区块高度"
// @Success      200    {object}  []model.Reward
// @Failure      400    {object}  service.ErrRes
// @Router       /reward/{block} [get]
func blockReward(c *gin.Context) {
	block := c.Param("block")

	data, err := service.BlockRewards(block)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
