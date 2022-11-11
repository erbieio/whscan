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
	number, cache, taskNum := service.TotalBlock(), make(map[Uint64]*Parsed), int64(0)
	log.Printf("using %v coroutines, starting data analysis from %v block\n", thread, number)
	for {
		max, err := client.BlockNumber()
		if err != nil {
			log.Printf("get block height error: %v\n", err)
		}
		if max == number {
			service.UpdateComSNFT = true
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
			taskNum, cache[parsed.Number] = taskNum-1, parsed
			for i := service.TotalBlock(); cache[i] != nil; {
				if badBlocks, err := service.BlockInsert(cache[i]); err != nil {
					log.Printf("%v block write error: %v\n", number, err)
					time.Sleep(interval)
				} else if badBlocks == nil {
					delete(cache, i)
					i++
				} else {
					for {
						if head, err := checkHead(client, context.Background(), i-1, badBlocks); err == nil {
							if number, err = service.FixHead(head); err == nil {
								for ; taskNum > 0; taskNum-- {
									<-parsedCh
								}
								cache = make(map[Uint64]*Parsed)
								log.Printf("fork fallback %v, starting data analysis from %v blockn\n", i-number, number)
								break
							}
						}
						time.Sleep(interval)
					}
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
