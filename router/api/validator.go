package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/common/utils"
	"server/service"
)

// Validator ValidatorAPI
func Validator(e *gin.Engine) {
	e.GET("/validator/page", validators)
	e.GET("/validator/locations", locations)
	e.GET("/validator/last_msg", lastMsg)
}

// @Tags        validator
// @Summary     query validator list
// @Description Query validator's information
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} []model.Validator
// @Failure     400       {object} service.ErrRes
// @Router      /validator/page [get]
func validators(c *gin.Context) {
	req := struct {
		Order    string `form:"order"`
		Page     *int   `form:"page"`
		PageSize *int   `form:"page_size"`
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
	res, err := service.FetchValidator(page, size, req.Order)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        validator
// @Summary     query validator locations
// @Description Query validator's locations
// @Accept      json
// @Produce     json
// @Success     200 {object} []service.LocationRes
// @Failure     400 {object} service.ErrRes
// @Router      /validator/locations [get]
func locations(c *gin.Context) {
	res, err := service.FetchLocations()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        validator
// @Summary     query validator msg list
// @Description Query validator's last msg list
// @Accept      json
// @Produce     json
// @Success     200 {object} []service.Msg
// @Router      /validator/last_msg [get]
func lastMsg(c *gin.Context) {
	c.JSON(http.StatusOK, service.GetLastMsg())
}
