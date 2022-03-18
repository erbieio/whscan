package main

import (
	"fmt"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	_ "server/docs"
	"server/monitor"
	"server/routers"
)

// @title Gin swagger
// @version 1.0
// @description 区块链监控程序后端
// @host 192.168.1.237:3001
func main() {
	go monitor.SyncBlock()
	r := routers.Init()
	r.GET("/docs/*any", gs.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(":3001"); err != nil {
		fmt.Println("startup service failed, err: ", err)
	}
}
