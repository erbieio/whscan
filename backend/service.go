package backend

import (
	"context"
	"log"
	"time"

	"server/common/types"
	. "server/conf"
	"server/node"
	"server/service"
)

func Run() {
	interval := time.Duration(Interval) * time.Second
	ethClient, err := node.Dial(ChainUrl)
	if err != nil {
		panic(err)
	}
	go Loop(ethClient, interval)
}

func Loop(ec *node.Client, interval time.Duration) {
	number := service.TotalBlock()
	log.Printf("The query cache was initialized successfully, starting data analysis from the %v block", number)
	isDebug := ec.IsDebug()
	isWormholes := ec.IsWormholes()
	log.Printf("open debug api: %v, wormholes chain: %v", isDebug, isWormholes)
	if !isDebug {
		log.Printf("Not open debug api will result in some missing data")
	}
	for {
		err := HandleBlock(ec, number, isDebug, isWormholes)
		if err != nil {
			if err != node.NotFound {
				log.Printf("Sleep in block %v, error: %v", number, err)
			}
			time.Sleep(interval)
		} else {
			number++
		}
	}
}

func HandleBlock(ec *node.Client, number uint64, isDebug, isWormholes bool) error {
	ret, err := DecodeBlock(ec, context.Background(), types.Uint64(number), isDebug, isWormholes)
	if err != nil {
		return err
	}
	return service.BlockInsert(ret)
}
