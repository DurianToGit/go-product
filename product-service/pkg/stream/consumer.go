package stream

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductEventConsumer struct {
	rdb      *redis.Client
	stream   string
	group    string
	consumer string
}

func NewProductEventConsumer(rdb *redis.Client, consumerName string) *ProductEventConsumer {
	return &ProductEventConsumer{
		rdb:      rdb,
		stream:   "stream:product:event",
		group:    "product_event_group",
		consumer: consumerName,
	}
}

func (c *ProductEventConsumer) InitGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (c *ProductEventConsumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.consumer,
			Streams:  []string{c.stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Println("xreadgroup error:", err)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				// TODO: 处理业务
				log.Printf("consume msg id=%s values=%v", msg.ID, msg.Values)

				// ACK
				err = c.rdb.XAck(ctx, c.stream, c.group, msg.ID).Err()
				if err != nil {
					log.Printf("xack failed: stream=%s group=%s id=%s err=%v", c.stream, c.group, msg.ID, err)
				}
			}
		}
	}
}
