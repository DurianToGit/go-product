package service

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"log"
	"product-service/internal/domain"
	"product-service/internal/repository"
	"product-service/pkg/utils"
)

type OrderService struct {
	repo       repository.OrderRepository
	productSvc *ProductService
	gen        *utils.DistributedOrderGenerator
}

func NewOrderService(repo repository.OrderRepository, productService *ProductService) *OrderService {
	return &OrderService{
		repo:       repo,
		productSvc: productService,
		gen:        utils.NewDistributedOrderGenerator("order_"),
	}
}

func (s *OrderService) Create(ctx context.Context, userID, productID int64, count int, idemKey string) (*domain.Order, error) {
	var order *domain.Order

	// 尝试查找已存在的订单
	existingOrder, err := s.repo.GetByUserIdemKey(ctx, userID, idemKey)
	if err == nil && existingOrder != nil {
		// 如果已存在，直接返回
		return existingOrder, nil
	}

	// 如果不是记录不存在的错误，则返回错误
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建新订单
	order = &domain.Order{
		OrderNo:   s.gen.GenerateOrderID(),
		UserID:    userID,
		ProductID: productID,
		Status:    domain.OrderStatusCreated,
		Count:     count,
		IdemKey:   idemKey,
	}

	result, err := s.repo.Create(ctx, order)
	if err != nil {
		log.Printf("创建订单失败:%v", err)
		return nil, err
	}

	// 扣减库存
	_, err = s.productSvc.DeductStockSeckill(ctx, productID, int64(count), userID)
	if err != nil {
		derr := s.repo.Delete(ctx, result.ID)
		if derr != nil {
			log.Printf("扣库存失败[%v]后，删除订单失败:%v", err, derr)
			return nil, derr
		}
		return nil, err
	}
	return result, nil
}
