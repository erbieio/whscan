package main

import (
	"log"

	"server/backend"
	"server/conf"
	"server/router"
)

// @title       block explorer API
// @version     1.0
// @description Block browser back-end interface, parses data from the blockchain, provides information retrieval services for blocks, transactions, NFT, SNFT, validators, and rewards
func main() {
	if err := backend.Run(conf.ChainUrl, conf.Thread, conf.Interval); err != nil {
		log.Printf("Backend failed to run： %v\n", err)
	}
	if err := router.Run(conf.ServerAddr); err != nil {
		log.Printf("Server failed to run： %v\n", err)
	}
}
