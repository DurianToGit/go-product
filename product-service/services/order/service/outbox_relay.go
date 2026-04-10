package service

import (
	"context"
	"fmt"
	"product-service/services/order/domain"
	"strconv"

	"go.uber.org/zap"

	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	mysqlOrder "product-service/services/order/repository/mysql"
	orderModel "product-service/services/order/repository/mysql/model"
)

type OutboxRelay struct {
	outboxRepo *mysqlOrder.OutboxRepository
}

func NewOutboxRelay(outboxRepo *mysqlOrder.OutboxRepository) *OutboxRelay {
	return &OutboxRelay{outboxRepo: outboxRepo}
}

func (r *OutboxRelay) RunOnce(ctx context.Context, limit int) (int64, error) {
	events, err := r.outboxRepo.FindPending(ctx, limit)
	if err != nil {
		return 0, err
	}

	var successCnt int64
	for _, evt := range events {
		if err = r.publishOne(ctx, evt); err != nil {
			logger.L().Error("outbox publish failed",
				zap.Uint64("outbox_id", evt.ID),
				zap.String("event_type", evt.EventType),
				zap.Error(err),
			)
			err = r.outboxRepo.IncrementRetry(ctx, int64(evt.ID), err.Error())
			if err != nil {
				logger.L().Error("outbox increment retry failed",
					zap.Uint64("outbox_id", evt.ID),
					zap.Error(err),
				)
			}
			continue
		}

		if err = r.outboxRepo.MarkSent(ctx, int64(evt.ID)); err != nil {
			logger.L().Error("mark outbox sent failed",
				zap.Uint64("outbox_id", evt.ID),
				zap.Error(err),
			)
		} else {
			successCnt += 1
		}
	}

	return successCnt, nil
}

func (r *OutboxRelay) publishOne(ctx context.Context, evt *orderModel.OutboxEventModel) error {
	if kafka.Client == nil {
		return fmt.Errorf("kafka client is nil")
	}
	key := []byte(strconv.FormatInt(int64(evt.ID), 10))
	payload := evt.Payload
	switch evt.EventType {
	case domain.OutboxEventTypeOrderPaid:
		return kafka.PublishRaw(ctx, kafka.TopicOrderPaid, key, payload)
	case domain.OutboxEventTypeOrderCanceled:
		return kafka.PublishRaw(ctx, kafka.TopicOrderCanceled, key, payload)
	default:
		return fmt.Errorf("unsupported outbox event type: %s", evt.EventType)
	}
}
