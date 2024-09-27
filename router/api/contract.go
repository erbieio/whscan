package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/common/utils"
	"server/service"
	"strconv"
)

// Contract ContractAPI
func Contract(e *gin.Engine) {
	e.GET("/contract/page", pageContract)
	e.GET("/contract/:addr", getContract)
	e.GET("/contract/holders/page", pageHolders)
}

// @Tags        contract
// @Summary     query top contracts
// @Description set the contract ranking according to create time
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.ContractsRes
// @Failure     400       {object} service.ErrRes
// @Router      /contract/page [get]
func pageContract(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	ctype := c.Query("ctype")
	ty, err := strconv.Atoi(ctype)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}
	res, err := service.FetchContracts(ty, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        Contract
// @Summary     query one contract
// @Description Query the contract information of the specified address
// @Accept      json
// @Produce     json
// @Param       addr path     string true "contract address"
// @Success     200  {object} model.Contract
// @Failure     400  {object} service.ErrRes
// @Router      /contract/{addr} [get]
func getContract(c *gin.Context) {
	res, err := service.GetContract(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        holders
// @Summary     query holders
// @Description set the holders ranking according to amount
// @Accept      json
// @Produce     json
// @Param       order     query    string false "sort by conditions, Support database order statement"
// @Param       page      query    string false "Page, default 1"
// @Param       page_size query    string false "Page size, default 10"
// @Success     200       {object} service.HoldersRes
// @Failure     400       {object} service.ErrRes
// @Router      /contract/holders/page [get]
func pageHolders(c *gin.Context) {
	page, size := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	addr := c.Query("addr")

	res, err := service.FetchHolders(addr, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
