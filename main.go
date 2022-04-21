package main

import (
	"log"

	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"server/backend"
	. "server/conf"
	_ "server/docs"
	"server/router"
)

// @title        区块浏览器API
// @version      1.0
// @description  区块浏览器后端接口，从区块链解析数据，提供区块、交易、NFT、SNFT、NFT合集、交易所的信息检索服务
func main() {
	backend.Run()
	r := router.Init()
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(ServerAddr); err != nil {
		log.Println("服务器运行失败，", err)
	}
}
