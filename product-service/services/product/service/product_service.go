package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"product-service/internal/config"
	"product-service/internal/errno"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/pkg/redislock"
	"product-service/pkg/seckill"
	"product-service/pkg/stream"
	"product-service/services/product/domain"
	"product-service/services/product/dto"
	"product-service/services/product/repository"
	"time"
)

// ProductService provides:
// - CreateProduct
// - SearchProducts (cached + rate limited)
// - GetProductDetail
// - PrewarmProductStock
// - DeductStockSeckill (Redis atomic)
// - DeductStockAtomic (DB final consistency via worker)

type ProductService struct {
	repo                 repository.ProductRepository
	locker               *redislock.Locker
	productEventProducer *stream.ProductEventProducer
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo:                 repo,
		locker:               redislock.New(redis.Client),
		productEventProducer: stream.NewProductEventProducer(redis.Client),
	}
}

func (s *ProductService) GetProducts(ctx context.Context, q *dto.ProductQuery) ([]*domain.Product, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *ProductService) CreateProduct(ctx context.Context, p *domain.Product) error {
	return s.repo.Create(ctx, p)
}

func (s *ProductService) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProductService) GetStock(ctx context.Context, id int64) (int64, error) {
	product, err := s.repo.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return product.Stock, nil
}

func (s *ProductService) PrewarmProductStock(ctx context.Context, id int64) error {
	data, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	err = redis.Do(ctx, func() error {
		return redis.Client.Set(ctx, rediskeys.ProductStockKey(id), data.Stock, 0).Err()
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ProductService) DeductStockSeckill(ctx context.Context, productId int64, count, userId int64, idemKey string) (int64, error) {
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
			logger.L().Error("publish product event failed", zap.Error(err2))
			return 0, errno.ErrDependencyUnavailable
		}
		return remain, nil
	}
}

func (s *ProductService) RestoreStock(ctx context.Context, productId int64, count int64, source string) error {
	logger.L().Info("恢复库存", zap.Int64("product_id", productId), zap.Int64("count", count), zap.String("source", source))
	return s.repo.RestoreStock(ctx, productId, count)
}

func (s *ProductService) ConsumeStockDeductEvent(ctx context.Context, stream, msgID string, productID, count int64, eventType string) error {
	logger.L().Info("消费库存扣减事件", zap.Int64("product_id", productID), zap.Int64("count", count), zap.String("event_type", eventType))
	return s.repo.ConsumeStockDeductEvent(ctx, stream, msgID, productID, count, eventType)
}

func (s *ProductService) ConsumeRestockDeductEvent(ctx context.Context, stream, msgID string, productID, count int64, eventType string) error {
	logger.L().Info("消费恢复库存事件", zap.Int64("product_id", productID), zap.Int64("count", count), zap.String("event_type", eventType))
	return s.repo.ConsumeRestockDeductEvent(ctx, stream, msgID, productID, count, eventType)
}
