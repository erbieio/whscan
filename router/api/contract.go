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
	e.GET("/contract/total_num", getContractTotalNum)
	e.GET("/contract/token_total_num", getTokenTotalNum)
	e.GET("/contract/nft_total_num", getNftTotalNum)
	e.GET("/contract/transfer_num/:addr", getTransferNum)
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

// @Tags        contract total number
// @Summary     query contract total number
// @Description query contract total number
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /contract/total_num [get]
func getContractTotalNum(c *gin.Context) {
	res, err := service.GetContractTotalNum()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        token total number
// @Summary     query token total number
// @Description query token total number
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /contract/token_total_num [get]
func getTokenTotalNum(c *gin.Context) {
	res, err := service.GetTokenTotalNum()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        nft total number
// @Summary     query nft total number
// @Description query nft total number
// @Accept      json
// @Produce     json
// @Param
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /contract/nft_total_num [get]
func getNftTotalNum(c *gin.Context) {
	res, err := service.GetNftTotalNum()
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Tags        contract transfer number
// @Summary     query contract transfer number
// @Description query contract transfer number
// @Accept      json
// @Produce     json
// @Param		addr path     string true "contract address"
// @Success     200  {object} int64
// @Failure     400  {object} service.ErrRes
// @Router      /contract/transfer_num [get]
func getTransferNum(c *gin.Context) {
	res, err := service.GetTransferNum(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, service.ErrRes{ErrStr: err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
