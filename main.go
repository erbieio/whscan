package main

import (
	"log"

	"server/backend"
	"server/conf"
	"server/router"
)

// @title       block explorer API
// @version     1.0
// @description Block browser back-end interface, parses data from the blockchain, provides information retrieval services for blocks, transactions, NFT, SNFT, NFT collections, and exchanges
func main() {
	backend.Run(conf.ChainUrl, conf.Thread, conf.Interval)
	if err := router.Run(conf.ServerAddr); err != nil {
		log.Printf("Server failed to run： %v\n", err)
	}
}
