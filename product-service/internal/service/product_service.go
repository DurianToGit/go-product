package service

import (
	"context"
	"log"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/repository"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/pkg/redislock"
	"product-service/pkg/seckill"
	"product-service/pkg/stream"
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

func (s *ProductService) GetProductWithCreator(ctx context.Context, id int64) (*domain.Product, *domain.User, error) {
	return s.repo.GetWithCreator(ctx, id)
}

func (s *ProductService) PrewarmProductStock(ctx context.Context, id int64) error {
	data, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	err = redis.Client.Set(ctx, rediskeys.ProductStockKey(id), data.Stock, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *ProductService) DeductStockSeckill(ctx context.Context, productId int64, count, userId int64, idemKey string) (int64, error) {

	remain, err := seckill.DeductStockLua(ctx, redis.Client, productId, count)
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
		err2 := s.productEventProducer.PublishOnce(ctx2, onceKey, map[string]any{
			"product_id": productId,
			"user_id":    userId,
			"count":      count,
			"event_type": domain.ProductEventTypeStockDeducted,
			"source":     "seckill",
		}, 30*time.Minute)
		log.Printf("publish product event, product_id=%d,count=%d,user_id=%d", productId, count, userId)
		if err2 != nil {
			log.Printf("publish product event failed, err=%v, product_id=%d", err2, productId)
		}
		return remain, nil
	}
}
