package stream

import (
	"context"
	"go.uber.org/zap"
	"product-service/pkg/logger"
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

		// 1. 先处理 pending 太久的消息
		c.consumePending(ctx, handler)

		// 2. 再消费新消息
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
			logger.L().Error("xreadgroup error",
				zap.String("stream", c.stream),
				zap.String("group", c.group),
				zap.String("consumer", c.consumer),
				zap.Error(err),
			)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				logger.L().Info("consume msg",
					zap.String("stream", s.Stream),
					zap.String("msg_id", msg.ID),
				)
				herr := handler(ctx, msg)
				if herr != nil {
					logger.L().Error("handle msg failed",
						zap.String("msg_id", msg.ID),
						zap.Error(herr),
					)
					// 不 ACK，留 pending
					continue
				}

				// ACK
				aerr := c.rdb.XAck(ctx, c.stream, c.group, msg.ID).Err()
				if aerr != nil {
					logger.L().Error("xack failed",
						zap.String("stream", c.stream),
						zap.String("group", c.group),
						zap.String("msg_id", msg.ID),
						zap.Error(aerr),
					)
				}
			}
		}
	}
}

func (c *ProductEventConsumer) consumePending(ctx context.Context, handler Handler) {
	msgs, _, err := c.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   c.stream,
		Group:    c.group,
		Consumer: c.consumer,
		MinIdle:  30 * time.Second,
		Start:    "0-0",
		Count:    10,
	}).Result()

	if err != nil {
		if err != redis.Nil {
			logger.L().Error("xautoclaim failed",
				zap.String("stream", c.stream),
				zap.String("group", c.group),
				zap.String("consumer", c.consumer),
				zap.Error(err),
			)
		}
		return
	}

	for _, msg := range msgs {
		logger.L().Info("处理待处理消息",
			zap.String("stream", c.stream),
			zap.String("group", c.group),
			zap.String("consumer", c.consumer),
			zap.String("msg_id", msg.ID),
		)

		herr := handler(ctx, msg)
		if herr != nil {
			logger.L().Error("处理待定消息失败",
				zap.String("msg_id", msg.ID),
				zap.Any("values", msg.Values),
				zap.Error(herr),
			)
			// 不 ACK，留待后续继续重试
			continue
		}

		if err := c.rdb.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
			logger.L().Error("xack pending msg failed",
				zap.String("stream", c.stream),
				zap.String("group", c.group),
				zap.String("msg_id", msg.ID),
				zap.Error(err),
			)
		}
	}
}
