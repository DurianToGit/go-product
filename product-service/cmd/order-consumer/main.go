package main

import (
	"context"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/client/productclient"
	"product-service/pkg/db"
	"product-service/pkg/event"
	"product-service/pkg/grpcx"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/pb/orderpb"
	"product-service/pkg/redis"
	"product-service/services/order/repository"
	orderMysql "product-service/services/order/repository/mysql"
	"sync"
	"syscall"
)

// 订单消费者 使用 kafka

var mySQL *gorm.DB
var productClient productclient.Client

func main() {
	cfg := bootstrap.BaseInit()

	logger.InitFromEnv("order-consumer")
	defer logger.Sync()
	logger.L().Info("启动订单消费者")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL = db.InitMySQL(dsn)
	_ = redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	// 连接到 gRPC 服务端
	conn, err := grpc.Dial(
		cfg.App.Product.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		grpc.WithStreamInterceptor(grpcx.StreamClientLoggingInterceptor()),
	)
	if err != nil {
		logger.L().Fatal("连接 gRPC 服务失败", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
		if err != nil {
			logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
		}
	}(conn)

	productClient = productclient.NewGRPCClient(conn)
	// productRepo = mysql.NewProductRepository(mySQL)
	consumerLogRepo := orderMysql.NewEventConsumeLogRepository(mySQL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go consumeOrderCanceled(ctx, cfg.Kafka.Addr, &wg, consumerLogRepo)
	go consumeOrderPaid(ctx, cfg.Kafka.Addr, &wg, consumerLogRepo)

	// 监听系统信号，实现优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.L().Info("收到退出信号，正在关闭服务...", zap.String("signal", sig.String()))
		cancel()
	}()

	wg.Wait()
	logger.L().Info("所有消费者已停止，服务退出")
}

func consumeOrderPaid(ctx context.Context, broker []string, wg *sync.WaitGroup, logRepo repository.EventConsumeLogRepository) {
	defer wg.Done()
	consumer := kafka.NewConsumer(broker, kafka.TopicOrderPaid, kafka.GroupOrderPaid)
	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()
		if err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
	}(consumer)

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return handleOrderPaid(ctx, msg, logRepo)
	})
	if err != nil {
		logger.L().Error("消费 order.paid 失败", zap.Error(err))
	}
}

func handleOrderPaid(ctx context.Context, msg kafkago.Message, logRepo repository.EventConsumeLogRepository) error {
	pbEvt := &orderpb.OrderPaidEvent{}
	err := proto.Unmarshal(msg.Value, pbEvt)
	if err != nil {
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
	ok, err := logRepo.TryConsume(ctx, eventID, kafka.GroupOrderPaid)
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

	if err = processOrderPaid(ctx, evt); err != nil {
		return fmt.Errorf("process order paid failed: %w", err)
	}
	return nil
}

// 订单支付事件处理
func processOrderPaid(ctx context.Context, evt event.OrderPaidEvent) error {
	// TODO:
	logger.L().Info("处理订单支付事件：", zap.Any("data", evt))
	// 1. 加积分
	logger.L().Info("增加积分")
	// 2. 发通知
	logger.L().Info("发送通知")
	// 3. 生成发货单
	logger.L().Info("生成发货单")
	// ...更多业务处理
	return nil
}

func consumeOrderCanceled(ctx context.Context, broker []string, wg *sync.WaitGroup, logRepo repository.EventConsumeLogRepository) {
	defer wg.Done()
	consumer := kafka.NewConsumer(broker, kafka.TopicOrderCanceled, kafka.GroupOrderCanceled)
	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()
		if err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
	}(consumer)

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return handleOrderCanceled(ctx, msg, logRepo)
	})
	if err != nil {
		logger.L().Error("消费 order.canceled 失败", zap.Error(err))
	}
}

func handleOrderCanceled(ctx context.Context, msg kafkago.Message, logRepo repository.EventConsumeLogRepository) error {
	pbEvt := orderpb.OrderCanceledEvent{}
	if err := proto.Unmarshal(msg.Value, &pbEvt); err != nil {
		return fmt.Errorf("unmarshal order canceled event failed: %w", err)
	}
	evt := event.OrderCanceledEvent{
		OrderID:    pbEvt.OrderId,
		UserID:     pbEvt.UserId,
		ProductID:  pbEvt.ProductId,
		Count:      pbEvt.Count,
		Reason:     pbEvt.Reason,
		CanceledAt: pbEvt.CanceledAt,
	}
	logger.L().Info("收到取消订单事件",
		zap.Int64("order_id", evt.OrderID),
		zap.Int64("product_id", evt.ProductID),
		zap.Int64("count", evt.Count),
		zap.String("reason", evt.Reason),
	)
	eventID := string(msg.Key)
	if eventID == "" {
		return fmt.Errorf("empty event id")
	}

	ok, err := logRepo.TryConsume(ctx, eventID, kafka.GroupOrderCanceled)
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
	if err = productClient.RestoreStock(ctx, evt.ProductID, evt.Count, "cancel_order"); err != nil {
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
