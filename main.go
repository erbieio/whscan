package main

import (
	"log"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"server/backend"
	. "server/conf"
	_ "server/docs"
	"server/router"
)

// @title       block explorer API
// @version     1.0
// @description Block browser back-end interface, parses data from the blockchain, provides information retrieval services for blocks, transactions, NFT, SNFT, NFT collections, and exchanges
func main() {
	backend.Run()
	r := router.Init()
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := r.Run(ServerAddr); err != nil {
		log.Println("Server failed to run,", err)
	}
}
