package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Reward API
func Reward(e *gin.Engine) {
	e.GET("/reward", pageReward)
	e.GET("/reward/:block", blockReward)
}

// @Tags        reward
// @Summary     Query the reward list
// @Description query the reward list in reverse order
// @Accept      json
// @Produce     json
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.RewardsRes
// @Failure     400       {object} service.ErrRes
// @Router      /reward [get]
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

// @Tags        reward
// @Summary     query reward
// @Description specifies the block query reward
// @Accept      json
// @Produce     json
// @Param       block path     string true "Block height"
// @Success     200   {object} []service.Reward
// @Failure     400   {object} service.ErrRes
// @Router      /reward/{block} [get]
func blockReward(c *gin.Context) {
	block := c.Param("block")

	data, err := service.BlockRewards(block)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
