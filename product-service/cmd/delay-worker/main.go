package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"product-service/internal/bootstrap"
	"time"

	"product-service/pkg/delayqueue"
	"product-service/pkg/eventbus"
	"product-service/pkg/redis"
)

func main() {
	ctx := context.Background()

	cfg := bootstrap.BaseInit()
	rdb := redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	queue := delayqueue.NewDelayQueue(rdb, "delay:tasks")

	bus := eventbus.NewKafkaBus(cfg.Kafka.Addr)
	defer func(bus *eventbus.KafkaBus) {
		err := bus.Close()
		if err != nil {
			fmt.Println("close bus error:", err)
		}
	}(bus)

	for {
		now := time.Now().Unix()

		tasks, err := queue.PopReady(ctx, now, 10)
		if err != nil {
			log.Println("pop error:", err)
			time.Sleep(time.Second)
			continue
		}

		var successMembers []string

		for _, t := range tasks {
			var task delayqueue.Task
			if err := json.Unmarshal([]byte(t), &task); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}

			switch task.BizType {
			case "stock.deduct.retry":
				if err := bus.Publish(ctx, "stock.deduct.requested", "", task.Payload); err != nil {
					log.Println("publish error:", err)
					continue
				}
				successMembers = append(successMembers, t)

			case "order.auto_cancel":
				if err := bus.Publish(ctx, "order.cancelled", "", task.Payload); err != nil {
					log.Println("publish error:", err)
					continue
				}
				successMembers = append(successMembers, t)

			default:
				log.Println("unknown biz type:", task.BizType)
				continue
			}
		}

		if len(successMembers) > 0 {
			_ = queue.Remove(ctx, successMembers)
		}

		time.Sleep(time.Second)
	}
}
