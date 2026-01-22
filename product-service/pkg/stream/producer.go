package stream

import (
	"context"
	"fmt"
	"product-service/pkg/rediskeys"
	"time"

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

// PublishOnce: 幂等发布（SETNX 门闩）
func (p *ProductEventProducer) PublishOnce(ctx context.Context, onceKey string, payload map[string]any, ttl time.Duration) error {
	// 1) SETNX 门闩
	ok, err := p.rdb.SetNX(ctx, onceKey, 1, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		// 已发布过（或并发下别人拿到了锁），直接视为成功
		return nil
	}

	// 2) 真正 publish（复用你原 Publish）
	return p.Publish(ctx, payload)
}

// 工具：拼 key
func SideFxKeyDeduct(userID int64, idemKey string) string {
	return fmt.Sprintf("order:sidefx:deduct:%d:%s", userID, idemKey)
}
func SideFxKeyRestock(orderID int64) string {
	return fmt.Sprintf("order:sidefx:restock:%d", orderID)
}
