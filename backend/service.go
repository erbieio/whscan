package backend

import (
	"context"
	"log"
	"time"

	"server/common/model"
	"server/common/types"
	"server/node"
)

func Run(chainUrl string, thread int64, interval time.Duration) (err error) {
	client, stats, ctx := &node.Client{}, &model.Stats{}, context.Background()
	if client, err = node.Dial(chainUrl); err != nil {
		return
	}
	if stats, err = check(client, ctx); err != nil {
		return
	}
	go loop(client, ctx, stats, thread, interval)
	return
}

func loop(client *node.Client, ctx context.Context, stats *model.Stats, thread int64, interval time.Duration) {
	parsedCh := make(chan *model.Parsed, thread)
	cache := make(map[types.Long]*model.Parsed)
	number, taskCount := types.Long(stats.TotalBlock), int64(0)
	log.Printf("using %v coroutines, starting data analysis from %v block\n", thread, number)
	for {
		max, err := client.BlockNumber(ctx)
		if err != nil || (number > max && taskCount == 0) {
			if err != nil {
				log.Printf("get block height error: %v\n", err)
			} else {
				stats.Ready = true
			}
			time.Sleep(interval)
		}
		for number <= max || taskCount > 0 {
			for ; number <= max && taskCount < thread; number++ {
				go func(number types.Long) {
					for {
						if parsed, err := decode(client, ctx, number); err != nil {
							log.Printf("%v block parsing error: %v\n", number, err)
							time.Sleep(10 * interval)
						} else {
							parsedCh <- parsed
							break
						}
					}
				}(number)
				taskCount++
			}
			parsed := <-parsedCh
			taskCount, cache[parsed.Number] = taskCount-1, parsed
			for newHead := types.Long(stats.TotalBlock); cache[newHead] != nil; {
				if head, err := write(client, ctx, cache[newHead]); err != nil {
					log.Printf("%v block write error: %v\n", newHead, err)
					time.Sleep(10 * interval)
				} else if head == newHead {
					delete(cache, newHead)
					newHead++
				} else {
					for ; taskCount > 0; taskCount-- {
						<-parsedCh
					}
					number, cache = head+1, make(map[types.Long]*model.Parsed)
					log.Printf("the header %v fork falls back to %v\n", newHead-1, head)
					break
				}
			}
		}
	}
}
