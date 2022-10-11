package backend

import (
	"context"
	"log"
	"sync"
	"time"

	. "server/common/model"
	. "server/common/types"
	. "server/conf"
	"server/node"
	"server/service"
)

var task = &sync.WaitGroup{}
var headCh = make(chan Uint64)
var runCh = make(chan int)

func Run() {
	client, err := node.Dial(ChainUrl)
	if err != nil {
		panic(err)
	}
	number := Uint64(service.TotalBlock())
	isDebug := client.IsDebug()
	isWormholes := client.IsWormholes()
	log.Printf("open debug api: %v, wormholes chain: %v", isDebug, isWormholes)
	if !isDebug {
		log.Printf("Not open debug api will result in some missing data")
	}
	taskCh := make(chan Uint64, Thread)
	parsedCh := make(chan *Parsed, 2*Thread)
	go writeLoop(number, parsedCh)
	go decodeLoop(client, taskCh, parsedCh, isDebug, isWormholes)
	go dispatchLoop(client, number, taskCh)
	log.Printf("Start data analysis from %v block using %v coroutines\n", Thread, number)
}

func dispatchLoop(client *node.Client, number Uint64, taskCh chan<- Uint64) {
	for {
		max, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("Error getting block height: %v\n", err)
		}
		if err != nil || number >= max {
			time.Sleep(Interval)
		}
		for ; number < max; number++ {
			select {
			case head := <-headCh:
				task.Wait()
				runCh <- 0
				number = head
			case taskCh <- number:
				task.Add(1)
			}
		}
	}
}

func decodeLoop(client *node.Client, taskCh <-chan Uint64, parsedCh chan<- *Parsed, isDebug, isWormholes bool) {
	for i := int64(0); i < Thread; i++ {
		go func() {
			for number := range taskCh {
				for {
					parsed, err := decode(client, context.Background(), number, isDebug, isWormholes)
					if err != nil {
						log.Printf("%v block parsing error: %v\n", number, err)
						time.Sleep(10 * Interval)
					} else {
						parsedCh <- parsed
						task.Done()
						break
					}
				}
			}
		}()
	}
}

func writeLoop(number Uint64, parsedCh <-chan *Parsed) {
	cache := make(map[Uint64]*Parsed)
	for parsed := range parsedCh {
		cache[parsed.Number] = parsed
		for cache[number] != nil {
			head, err := service.BlockInsert(cache[number])
			if err != nil {
				log.Printf("%v block write error: %v\n", number, err)
				time.Sleep(Interval)
			} else {
				if head != cache[number].Number {
					headCh <- head
					for {
						select {
						case <-parsedCh:
						case <-runCh:
							goto end
						}
					}
				end:
					number = head
					cache = make(map[Uint64]*Parsed)
				} else {
					delete(cache, number)
				}
				number++
			}
		}
	}
}
