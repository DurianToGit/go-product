package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"product-service/internal/client/productclient"
	"product-service/internal/config"
	"product-service/internal/errno"
	"product-service/pkg/event"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	redisPkg "product-service/pkg/redis"
	"product-service/pkg/stream"
	"product-service/pkg/utils"
	"product-service/services/order/domain"
	"product-service/services/order/repository"
	"strings"
	"time"
)

type OrderService struct {
	Repo                 repository.OrderRepository
	Gen                  *utils.DistributedOrderGenerator
	productEventProducer *stream.ProductEventProducer
	productClient        productclient.Client
}

func NewOrderService(repo repository.OrderRepository, productClient productclient.Client) *OrderService {
	return &OrderService{
		Repo:                 repo,
		Gen:                  utils.NewDistributedOrderGenerator("order_"),
		productEventProducer: stream.NewProductEventProducer(redisPkg.Client),
		productClient:        productClient,
	}
}

func (s *OrderService) Create(ctx context.Context, userID, productID int64, count int, idemKey string) (*domain.Order, error) {
	tr := otel.Tracer("order-service")
	ctx, span := tr.Start(ctx, "OrderService.Create")
	defer span.End()

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
		logger.L().Error("查找订单失败", zap.Error(err))
		return nil, err
	}
	product, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		logger.L().Error("获取商品信息失败", zap.Error(err))
		return nil, err
	}
	if product.Stock < int64(count) {
		return nil, errno.OrderErrNotEnoughStock
	}

	amount := product.Price * int64(count)
	// 创建新订单
	order = &domain.Order{
		OrderNo:   s.Gen.GenerateOrderID(),
		UserID:    userID,
		ProductID: productID,
		Status:    domain.OrderStatusCreated,
		Count:     count,
		Amount:    amount,
		IdemKey:   idemKey,
	}
	logger.L().Info("创建订单", zap.Any("order", order))

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
		logger.L().Error("查找订单失败", zap.Error(err))
		return nil, err
	}

	// 扣减库存
	_, err = s.DeductStockSeckill(ctx, productID, int64(count), userID, idemKey)
	if err != nil {
		derr := s.Repo.Delete(ctx, result.ID)
		if derr != nil {
			logger.L().Error("扣库存失败，删除订单失败。", zap.Error(err), zap.Error(derr))
			return nil, derr
		}
		return nil, err
	}
	return result, nil
}

func (s *OrderService) CancelExpired(ctx context.Context, now time.Time, timeout time.Duration) (int64, error) {
	tr := otel.Tracer("order-service")
	ctx, span := tr.Start(ctx, "OrderService.CancelExpired")
	defer span.End()
	deadline := now.Add(-timeout)
	num, data, err := s.Repo.CancelExpired(ctx, deadline)
	if err != nil {
		logger.L().Error("取消订单失败", zap.Error(err))
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
			logger.L().Error("恢复库存流发送失败", zap.Error(err2))
			return num, errno.ErrDependencyUnavailable
		}
	}
	return num, nil
}

func (s *OrderService) Get(ctx context.Context, id int64) (*domain.Order, error) {
	return s.Repo.Get(ctx, id)
}

func (s *OrderService) Pay(ctx context.Context, orderID int64) error {
	// 查询订单信息（构建事件）
	order, err := s.Repo.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errno.OrderNotFound
	}
	if order.Status != domain.OrderStatusCreated {
		return errno.OrderStatusInvalid
	}
	// 更新订单状态
	err = s.Repo.MarkPaid(ctx, orderID)
	if err != nil {
		return err
	}

	// 3. 发布事件
	evt := event.OrderPaidEvent{
		OrderID: order.ID,
		UserID:  order.UserID,
		Amount:  order.Amount,
		PaidAt:  time.Now().Unix(),
	}

	err = kafka.PublishOrderPaid(ctx, evt) // D45先不处理失败重试
	if err != nil {
		logger.L().Error("发布订单支付事件失败", zap.Error(err))
		// return err
	}

	return nil
}
