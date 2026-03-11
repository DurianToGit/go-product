package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"product-service/internal/config"
	"product-service/internal/errno"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/pkg/seckill"
	"product-service/pkg/stream"
	"product-service/services/order/domain"
	"time"
)

func (s *OrderService) DeductStockSeckill(ctx context.Context, productId int64, count, userId int64, idemKey string) (int64, error) {
	tr := otel.Tracer("product-service")
	ctx, span := tr.Start(ctx, "ProductService.DeductStockSeckill")
	defer span.End()

	remain, err := seckill.DeductStockLua(ctx, redis.Client, productId, count)
	cfg := config.GetRuntimeConfig()
	if err != nil {
		return 0, err
	}
	switch remain {
	case -2:
		return 0, errno.ProductErrSeckillStockNotInit
	case -1:
		return 0, errno.ProductErrSeckillStockNotEnough
	default:
		ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		onceKey := stream.SideFxKeyDeduct(userId, idemKey)
		ctx2, span2 := otel.Tracer("stream").Start(ctx2, "stream.publish")
		defer span2.End()
		data := map[string]any{
			"product_id": productId,
			"user_id":    userId,
			"count":      count,
			"event_type": domain.ProductEventTypeStockDeducted,
			"source":     "seckill",
		}
		err2 := s.productEventProducer.PublishOnce(ctx2, onceKey, data, time.Duration(cfg.OrderCancelTimeoutSec)*time.Second)
		logger.L().Info("publish product event", zap.Any("data", data))
		if err2 != nil {
			logger.L().Error("publish product event failed", zap.Error(err2), zap.Int64("product_id", productId))
			return 0, errno.ErrDependencyUnavailable
		}
		return remain, nil
	}
}
