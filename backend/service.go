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
	chainId, err := client.ChainId(context.Background())
	if err != nil {
		panic(err)
	}
	genesis, err := client.Genesis(context.Background())
	if err != nil {
		panic(err)
	}
	if err = service.CheckStats(chainId, genesis); err != nil {
		panic(err)
	}
	isDebug := client.IsDebug()
	isWormholes := client.IsWormholes()
	log.Printf("open debug api: %v, wormholes chain: %v", isDebug, isWormholes)
	if !isDebug {
		log.Printf("not open debug api will result in some missing data\n")
	}
	taskCh := make(chan Uint64, thread)
	parsedCh := make(chan *Parsed, thread)
	go taskLoop(client, thread, interval, taskCh, parsedCh, isDebug, isWormholes)
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
		if err != nil || (max <= number && taskNum == 0) {
			time.Sleep(interval)
		}
		for number <= max || taskNum > 0 {
			for number <= max && taskNum < thread {
				taskCh <- number
				taskNum++
				number++
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
						err = service.FixHead(head)
						if err != nil {
							break
						}
						for ; taskNum > 0; taskNum-- {
							<-parsedCh
						}
						number = head.Number + 1
						cache = make(map[Uint64]*Parsed)
						log.Printf("fork fallback, starting data analysis from %v blockn\n", number)
					}
					delete(cache, i)
				} else {
					break
				}
			}
		}
	}
}

func taskLoop(client *node.Client, thread int64, interval time.Duration, taskCh <-chan Uint64, parsedCh chan<- *Parsed, isDebug, isWormholes bool) {
	for ; thread > 0; thread-- {
		go func() {
			for number := range taskCh {
				for {
					parsed, err := decode(client, context.Background(), number, isDebug, isWormholes)
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
