package backend

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"server/common/types"
	. "server/conf"
	"server/ethclient"
	"server/service"
)

func Run() {
	interval := time.Duration(Interval) * time.Second
	ethClient, err := ethclient.Dial(ChainUrl)
	if err != nil {
		panic(err)
	}
	go Loop(ethClient, interval)
}

func Loop(ec *ethclient.Client, interval time.Duration) {
	number := service.TotalBlock()
	log.Printf("查询缓存初始化成功, 从%v区块开始数据分析", number)
	for {
		err := HandleBlock(ec, number)
		if err != nil {
			if err != ethereum.NotFound {
				log.Printf("在%v区块休眠, 错误：%v", number, err)
			}
			time.Sleep(interval)
		} else {
			number++
		}
	}
}

func HandleBlock(ec *ethclient.Client, number uint64) error {
	ret, err := ec.DecodeBlock(context.Background(), types.Uint64(number))
	if err != nil {
		return err
	}
	return service.BlockInsert(ret)
}
