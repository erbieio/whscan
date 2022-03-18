package extra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/big"
	"net/http"
	"server/ethhelper"
	"server/log"
	"time"
)

func Routers(e *gin.Engine) {
	e.GET("/extra/checkAuth", checkAuth)
}

// @Tags  其他接口
// @Summary 查询交易所状态
// @Description 查询交易所状态
// @Accept json
// @Produce json
// @Param body body CheckAuthReq true "body"
// @Success 200 {object} CheckAuthRes
// @Failure 400 {object} ErrRes
// @Router /extra/checkAuth [get]
func checkAuth(c *gin.Context) {
	var req CheckAuthReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}

	flag, balance, err := getAccountInfoFromGeth(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}
	res, err := ethhelper.CheckAuth(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, CheckAuthRes{Code: -1, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, CheckAuthRes{Code: 0, Msg: "ok", Data: CheckAuthResData{
		Status:           res,
		ExchangerFlag:    flag,
		ExchangerBalance: balance,
	}})
}
func getAccountInfoFromGeth(addr string) (bool, string, error) {
	type Params struct {
		JsonRpc string   `json:"jsonrpc"`
		Method  string   `json:"method"`
		Params  []string `json:"params"`
		Id      string   `json:"id"`
	}
	type Data struct {
		ExchangerFlag    bool    `json:"ExchangerFlag" `
		ExchangerBalance big.Int `json:"ExchangerBalance" `
	}
	type Ret struct {
		Data `json:"result" `
	}
	var p Params
	p.Params = append(p.Params, addr)
	p.Params = append(p.Params, "latest")
	p.JsonRpc = "2.0"
	p.Method = "eth_getAccountInfo"
	p.Id = "51888"
	contentType := "application/json"
	client := &http.Client{Timeout: 10 * time.Second}
	jsonStr, _ := json.Marshal(p)
	req, err := http.NewRequest("POST", "http://192.168.1.235:8561", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Info("GetAccountInfoFromGeth http.NewRequest err:", err)
		return false, "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("GetAccountInfoFromGeth http.NewRequest err:", err)
	}

	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	var r Ret
	err = json.Unmarshal(result, &r)

	if err != nil {
		fmt.Println("GetAccountInfoFromGeth json.Unmarshal err:", err)
		return false, "", err
	}
	return r.Data.ExchangerFlag, r.Data.ExchangerBalance.String(), err
}
