package stream

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type ProductEventProducer struct {
	rdb    *redis.Client
	stream string
}

func NewProductEventProducer(rdb *redis.Client) *ProductEventProducer {
	return &ProductEventProducer{
		rdb:    rdb,
		stream: "stream:product:event",
	}
}

func (p *ProductEventProducer) Publish(ctx context.Context, values map[string]any) error {
	_, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.stream,
		Values: values,
	}).Result()
	return err
}
