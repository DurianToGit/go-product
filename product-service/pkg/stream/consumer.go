package stream

import (
	"context"
	"log"
	"product-service/pkg/rediskeys"
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

type Handler func(ctx context.Context, msg redis.XMessage) error

func NewProductEventConsumer(rdb *redis.Client, consumerName string) *ProductEventConsumer {
	return &ProductEventConsumer{
		rdb:      rdb,
		stream:   rediskeys.ProductStreamKey,
		group:    "product_event_group",
		consumer: consumerName,
	}
}

func (c *ProductEventConsumer) InitGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err == nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (c *ProductEventConsumer) Consume(ctx context.Context, handler Handler) {
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
				herr := handler(ctx, msg)
				if herr != nil {
					log.Printf("handle msg failed: id=%s err=%v", msg.ID, herr)
					// 不 ACK，留 pending
					continue
				}

				// ACK
				aerr := c.rdb.XAck(ctx, c.stream, c.group, msg.ID).Err()
				if aerr != nil {
					log.Printf("xack failed: stream=%s group=%s id=%s err=%v", c.stream, c.group, msg.ID, aerr)
				}
			}
		}
	}
}
