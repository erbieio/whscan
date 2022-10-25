package backend

import (
	"context"
	"log"
	"time"

	. "server/common/model"
	. "server/common/types"
	"server/node"
	"server/service"
)

func Run(chainUrl string, thread int64, interval time.Duration) {
	client, err := node.Dial(chainUrl)
	if err != nil {
		panic(err)
	}
	chainId, err := client.ChainId()
	if err != nil {
		panic(err)
	}
	genesis, err := client.Genesis()
	if err != nil {
		panic(err)
	}
	if !service.CheckStats(chainId, genesis) {
		panic("Stored data and chain node information do not match")
	}
	if !client.IsDebug() || !client.IsWormholes() {
		panic("not open debug api or not exist wormholes api\n")
	}
	taskCh := make(chan Uint64, thread)
	parsedCh := make(chan *Parsed, thread)
	go taskLoop(client, thread, interval, taskCh, parsedCh)
	go mainLoop(client, thread, interval, taskCh, parsedCh)
}

func mainLoop(client *node.Client, thread int64, interval time.Duration, taskCh chan<- Uint64, parsedCh <-chan *Parsed) {
	number := service.TotalBlock()
	cache := make(map[Uint64]*Parsed)
	taskNum := int64(0)
	log.Printf("using %v coroutines, starting data analysis from %v blockn\n", thread, number)
	for {
		max, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("get block height error: %v\n", err)
		}
		if err != nil || (number > max && taskNum == 0) {
			time.Sleep(interval)
		}
		for number <= max || taskNum > 0 {
			for ; number <= max && taskNum < thread; number++ {
				taskCh <- number
				taskNum++
			}
			parsed := <-parsedCh
			taskNum--
			cache[parsed.Number] = parsed
			for i := service.TotalBlock(); cache[i] != nil; i++ {
				badBlocks, err := service.BlockInsert(cache[i])
				if err == nil {
					if badBlocks != nil {
						head, err := checkHead(client, context.Background(), i-1, badBlocks)
						if err != nil {
							break
						}
						if number, err = service.FixHead(head); err != nil {
							break
						}
						for ; taskNum > 0; taskNum-- {
							<-parsedCh
						}
						cache = make(map[Uint64]*Parsed)
						log.Printf("fork fallback, starting data analysis from %v blockn\n", number)
					}
					delete(cache, i)
				} else {
					time.Sleep(interval)
					log.Printf("write block error: %v\n", err)
					break
				}
			}
		}
	}
}

func taskLoop(client *node.Client, thread int64, interval time.Duration, taskCh <-chan Uint64, parsedCh chan<- *Parsed) {
	for ; thread > 0; thread-- {
		go func() {
			for number := range taskCh {
				for {
					parsed, err := decode(client, context.Background(), number)
					if err != nil {
						log.Printf("%v block parsing error: %v\n", number, err)
						time.Sleep(interval)
					} else {
						parsedCh <- parsed
						break
					}
				}
			}
		}()
	}
}
