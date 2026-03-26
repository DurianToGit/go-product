package main

import (
	"context"
	"encoding/json"
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

	for {
		now := time.Now().Unix()

		tasks, err := queue.PopReady(ctx, now, 10)
		if err != nil {
			log.Println("pop error:", err)
			time.Sleep(time.Second)
			continue
		}

		for _, t := range tasks {
			var task delayqueue.Task
			_ = json.Unmarshal([]byte(t), &task)

			switch task.BizType {

			case "stock.deduct.retry":
				// 重新投递 Kafka
				err := bus.Publish(ctx, "stock.deduct.requested", "", task.Payload)
				if err != nil {
					log.Println("publish error:", err)
					continue
				}

			case "order.auto_cancel":
				err := bus.Publish(ctx, "order.cancelled", "", task.Payload)
				if err != nil {
					log.Println("publish error:", err)
					continue
				}
			}
		}

		_ = queue.Remove(ctx, tasks)

		time.Sleep(time.Second)
	}
}
