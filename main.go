package main

import (
	"fmt"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	_ "server/docs"
	"server/monitor"
	"server/routers"
)

// @title        区块浏览器API
// @version      1.0
// @description  区块浏览器后端接口，从区块链解析数据，提供区块、交易、NFT、SNFT、NFT合集、交易所的信息检索服务
func main() {
	go monitor.SyncBlock()
	r := routers.Init()
	r.GET("/docs/*any", gs.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(":3001"); err != nil {
		fmt.Println("startup service failed, err: ", err)
	}
}
