package service

import (
	"context"
	"log"
	"product-service/internal/config"
	"product-service/internal/domain"
	"product-service/internal/errno"
	"product-service/internal/repository"
	"product-service/pkg/redis"
	"product-service/pkg/stream"
	"product-service/pkg/utils"
	"strings"
	"time"
)

type OrderService struct {
	Repo                 repository.OrderRepository
	ProductSvc           *ProductService
	Gen                  *utils.DistributedOrderGenerator
	productEventProducer *stream.ProductEventProducer
}

func NewOrderService(repo repository.OrderRepository, productService *ProductService) *OrderService {
	return &OrderService{
		Repo:                 repo,
		ProductSvc:           productService,
		Gen:                  utils.NewDistributedOrderGenerator("order_"),
		productEventProducer: stream.NewProductEventProducer(redis.Client),
	}
}

func (s *OrderService) Create(ctx context.Context, userID, productID int64, count int, idemKey string) (*domain.Order, error) {
	var order *domain.Order

	// 尝试查找已存在的订单
	existingOrder, err := s.Repo.GetByUserIdemKey(ctx, userID, idemKey)
	if err == nil && existingOrder != nil {
		// 如果已存在，但是信息不一致，则返回错误
		if existingOrder.ProductID != productID || existingOrder.Count != count {
			return nil, errno.OrderErrOrderAlreadyExist
		}
		return existingOrder, nil
	}

	// 如果不是记录不存在的错误，则返回错误
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		log.Printf("查找订单失败:%v", err)
		return nil, err
	}

	// 创建新订单
	order = &domain.Order{
		OrderNo:   s.Gen.GenerateOrderID(),
		UserID:    userID,
		ProductID: productID,
		Status:    domain.OrderStatusCreated,
		Count:     count,
		IdemKey:   idemKey,
	}

	result, err := s.Repo.Create(ctx, order)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			// 尝试查找已存在的订单
			existOrder, eerr := s.Repo.GetByUserIdemKey(ctx, userID, idemKey)
			if eerr == nil && existOrder != nil {
				// 如果已存在，直接返回
				return existOrder, nil
			}
		}
		log.Printf("创建订单失败:%v", err)
		return nil, err
	}

	// 扣减库存
	_, err = s.ProductSvc.DeductStockSeckill(ctx, productID, int64(count), userID, idemKey)
	if err != nil {
		derr := s.Repo.Delete(ctx, result.ID)
		if derr != nil {
			log.Printf("扣库存失败[%v]后，删除订单失败:%v", err, derr)
			return nil, derr
		}
		return nil, err
	}
	return result, nil
}

func (s *OrderService) CancelExpired(ctx context.Context, now time.Time, timeout time.Duration) (int64, error) {
	deadline := now.Add(-timeout)
	num, data, err := s.Repo.CancelExpired(ctx, deadline)
	if err != nil {
		log.Printf("取消订单失败:%v", err)
		return 0, err
	}
	if num == 0 {
		return 0, nil
	}
	cfg := config.GetRuntimeConfig()
	for _, order := range data {
		// 恢复库存
		onceKey := stream.SideFxKeyRestock(order.ID)
		err2 := s.productEventProducer.PublishOnce(ctx, onceKey, map[string]any{
			"product_id": order.ProductID,
			"count":      order.Count,
			"event_type": domain.ProductEventTypeRestockDeducted,
			"user_id":    order.UserID,
			"order_id":   order.ID,
			"source":     "cancelExpiredOrder",
		}, time.Duration(cfg.OrderCancelTimeoutSec)*time.Second)
		if err2 != nil {
			log.Printf("恢复库存流发送失败:%v", err2)
		}
	}
	return num, nil
}

func (s *OrderService) Get(ctx context.Context, id int64) (*domain.Order, error) {
	return s.Repo.Get(ctx, id)
}
