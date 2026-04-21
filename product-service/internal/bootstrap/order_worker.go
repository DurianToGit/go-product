package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis2 "github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"product-service/internal/client/productclient"
	"product-service/internal/config"
	"product-service/pkg/db"
	"product-service/pkg/event"
	"product-service/pkg/grpcx"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/pb/orderpb"
	"product-service/pkg/redis"
	orderRepository "product-service/services/order/repository"
	orderMysql "product-service/services/order/repository/mysql"
	orderService "product-service/services/order/service"
)

type OrderWorkerApp struct {
	cfg     *config.Config
	mysqlDB *gorm.DB
	redis   *redis2.Client

	productConn   *grpc.ClientConn
	productClient productclient.Client

	orderRepo      orderRepository.OrderRepository
	outboxRepo     *orderMysql.OutboxRepository
	consumeLogRepo orderRepository.EventConsumeLogRepository
	outboxRelay    *orderService.OutboxRelay

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func InitOrderWorkerApp() (*OrderWorkerApp, error) {
	cfg := BaseInit()

	// 资源清理函数列表，用于初始化失败时回滚
	var cleanup []func()

	defer func() {
		if cleanup != nil {
			for i := len(cleanup) - 1; i >= 0; i-- {
				cleanup[i]()
			}
		}
	}()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	cleanup = append(cleanup, func() {
		if sqlDB, err := mySQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.L().Error("关闭 MySQL 连接失败", zap.Error(err))
			}
		}
	})

	rdb := redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	cleanup = append(cleanup, func() {
		if err := rdb.Close(); err != nil {
			logger.L().Error("关闭 Redis 连接失败", zap.Error(err))
		}
	})

	// 初始化 gRPC 连接到 product 服务
	var productGrpcConn *grpc.ClientConn
	var grpcErr error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	productGrpcConn, grpcErr = grpc.DialContext(
		ctx,
		cfg.App.Product.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
	)
	cancel()
	if grpcErr != nil {
		logger.L().Error("初始化 grpc client 失败", zap.Error(grpcErr))
		return nil, grpcErr
	}
	cleanup = append(cleanup, func() {
		if err := productGrpcConn.Close(); err != nil {
			logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
		}
	})

	// 初始化 kafka client
	kafka.InitClient(cfg.Kafka.Addr)

	orderRepo := orderMysql.NewOrderRepository(mySQL)
	outboxRepo := orderMysql.NewOutboxRepository(mySQL)
	consumeLogRepo := orderMysql.NewEventConsumeLogRepository(mySQL)
	productClient := productclient.NewGRPCClient(productGrpcConn)
	outboxRelay := orderService.NewOutboxRelay(outboxRepo)

	cleanup = nil

	return &OrderWorkerApp{
		cfg:            cfg,
		mysqlDB:        mySQL,
		redis:          rdb,
		productConn:    productGrpcConn,
		productClient:  productClient,
		orderRepo:      orderRepo,
		outboxRepo:     outboxRepo,
		consumeLogRepo: consumeLogRepo,
		outboxRelay:    outboxRelay,
	}, nil
}

// Serve 启动 outbox relay ticker 和 kafka 消费者
func (a *OrderWorkerApp) Serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	// 启动 outbox relay ticker
	a.wg.Add(1)
	go a.runOutboxRelay(ctx)

	// 启动 kafka consumer: order.paid
	a.wg.Add(1)
	go a.consumeOrderPaid(ctx)

	// 启动 kafka consumer: order.canceled
	a.wg.Add(1)
	go a.consumeOrderCanceled(ctx)

	return nil
}

func (a *OrderWorkerApp) runOutboxRelay(ctx context.Context) {
	defer a.wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.L().Info("outbox relay stopped")
			return
		case <-ticker.C:
			cnt, err := a.outboxRelay.RunOnce(ctx, 100)
			if err != nil {
				logger.L().Error("outbox relay run once failed", zap.Error(err))
				continue
			}
			if cnt > 0 {
				logger.L().Info("outbox relay published", zap.Int64("count", cnt))
			}
		}
	}
}

func (a *OrderWorkerApp) consumeOrderPaid(ctx context.Context) {
	defer a.wg.Done()
	consumer := kafka.NewConsumer(a.cfg.Kafka.Addr, kafka.TopicOrderPaid, kafka.GroupOrderPaid)
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.L().Error("关闭 order.paid 消费者失败", zap.Error(err))
		}
	}()

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return a.handleOrderPaid(ctx, msg)
	})
	if err != nil && ctx.Err() == nil {
		logger.L().Error("消费 order.paid 失败", zap.Error(err))
	}
}

func (a *OrderWorkerApp) handleOrderPaid(ctx context.Context, msg kafkago.Message) error {
	pbEvt := &orderpb.OrderPaidEvent{}
	if err := proto.Unmarshal(msg.Value, pbEvt); err != nil {
		return fmt.Errorf("反序列化订单支付事件失败: %w", err)
	}

	eventID := string(msg.Key)
	if eventID == "" {
		return fmt.Errorf("empty event id")
	}

	evt := event.OrderPaidEvent{
		OrderID: pbEvt.OrderId,
		UserID:  pbEvt.UserId,
		Amount:  pbEvt.Amount,
		PaidAt:  pbEvt.PaidAt,
	}

	ok, err := a.consumeLogRepo.TryConsume(ctx, eventID, kafka.GroupOrderPaid)
	if err != nil {
		return fmt.Errorf("try consume order paid failed: %w", err)
	}
	if !ok {
		logger.L().Info("订单支付事件已处理过",
			zap.String("event_id", eventID),
			zap.Int64("order_id", evt.OrderID),
			zap.String("group", kafka.GroupOrderPaid),
		)
		return nil
	}

	// TODO: 处理订单支付业务逻辑（加积分、发通知、生成发货单等）
	logger.L().Info("处理订单支付事件",
		zap.Int64("order_id", evt.OrderID),
		zap.Int64("user_id", evt.UserID),
		zap.Int64("amount", evt.Amount),
	)
	return nil
}

func (a *OrderWorkerApp) consumeOrderCanceled(ctx context.Context) {
	defer a.wg.Done()
	consumer := kafka.NewConsumer(a.cfg.Kafka.Addr, kafka.TopicOrderCanceled, kafka.GroupOrderCanceled)
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.L().Error("关闭 order.canceled 消费者失败", zap.Error(err))
		}
	}()

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return a.handleOrderCanceled(ctx, msg)
	})
	if err != nil && ctx.Err() == nil {
		logger.L().Error("消费 order.canceled 失败", zap.Error(err))
	}
}

func (a *OrderWorkerApp) handleOrderCanceled(ctx context.Context, msg kafkago.Message) error {
	pbEvt := &orderpb.OrderCanceledEvent{}
	if err := proto.Unmarshal(msg.Value, pbEvt); err != nil {
		return fmt.Errorf("反序列化订单取消事件失败: %w", err)
	}

	eventID := string(msg.Key)
	if eventID == "" {
		return fmt.Errorf("empty event id")
	}

	evt := event.OrderCanceledEvent{
		OrderID:    pbEvt.OrderId,
		UserID:     pbEvt.UserId,
		ProductID:  pbEvt.ProductId,
		Count:      pbEvt.Count,
		Reason:     pbEvt.Reason,
		CanceledAt: pbEvt.CanceledAt,
	}

	ok, err := a.consumeLogRepo.TryConsume(ctx, eventID, kafka.GroupOrderCanceled)
	if err != nil {
		return fmt.Errorf("try consume order canceled failed: %w", err)
	}
	if !ok {
		logger.L().Info("订单取消事件已处理过",
			zap.String("event_id", eventID),
			zap.Int64("order_id", evt.OrderID),
			zap.String("group", kafka.GroupOrderCanceled),
		)
		return nil
	}

	if err = a.productClient.RestoreStock(ctx, evt.ProductID, evt.Count, "cancel_order"); err != nil {
		return fmt.Errorf("restore stock failed: %w", err)
	}

	logger.L().Info("订单取消事件消费成功",
		zap.String("event_id", eventID),
		zap.Int64("order_id", evt.OrderID),
		zap.Int64("product_id", evt.ProductID),
		zap.Int64("count", evt.Count),
	)
	return nil
}

// Close 优雅关闭应用资源
func (a *OrderWorkerApp) Close() error {
	var errs []error

	// 取消所有 goroutine
	if a.cancel != nil {
		a.cancel()
	}
	// 等待所有 goroutine 退出
	a.wg.Wait()

	// 关闭 gRPC 连接
	if a.productConn != nil {
		if err := a.productConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭 MySQL 连接
	if a.mysqlDB != nil {
		if sqlDB, err := a.mysqlDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 关闭 Redis 连接
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
