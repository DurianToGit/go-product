package stream

import (
	"context"
	"product-service/pkg/rediskeys"

	"github.com/redis/go-redis/v9"
)

type ProductEventProducer struct {
	rdb    *redis.Client
	stream string
}

func NewProductEventProducer(rdb *redis.Client) *ProductEventProducer {
	return &ProductEventProducer{
		rdb:    rdb,
		stream: rediskeys.ProductStreamKey,
	}
}

func (p *ProductEventProducer) Publish(ctx context.Context, values map[string]any) error {
	_, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.stream,
		Values: values,
	}).Result()
	return err
}
