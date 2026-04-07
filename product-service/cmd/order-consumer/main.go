package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"product-service/internal/bootstrap"
	"product-service/pkg/db"
	"product-service/pkg/event"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/services/product/repository/mysql"
	"sync"
)

var mySQL *gorm.DB

func main() {
	cfg := bootstrap.BaseInit()

	logger.InitFromEnv("order-consumer")
	defer logger.Sync()
	logger.L().Info("启动订单消费者")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL = db.InitMySQL(dsn)
	_ = redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	var wg sync.WaitGroup

	wg.Add(2)

	go consumeOrderCanceled(ctx, cfg.Kafka.Addr, &wg)
	go consumeOrderPaid(ctx, cfg.Kafka.Addr, &wg)

	wg.Wait()
	logger.L().Info("所有消费者已停止")
}

func consumeOrderPaid(ctx context.Context, broker []string, wg *sync.WaitGroup) {
	defer wg.Done()
	consumer := kafka.NewConsumer(broker, kafka.TopicOrderPaid, "order-paid-group")
	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()
		if err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
	}(consumer)

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return handleOrderPaid(msg)
	})
	if err != nil {
		logger.L().Error("消费 order.paid 失败", zap.Error(err))
	}
}

func handleOrderPaid(msg kafkago.Message) error {
	var evt event.OrderPaidEvent
	err := json.Unmarshal(msg.Value, &evt)
	if err != nil {
		logger.L().Error("解析订单支付事件失败：", zap.Error(err), zap.Any("data", evt))
		return fmt.Errorf("反序列化订单支付事件失败: %w", err)
	}

	logger.L().Info("增加积分：", zap.Any("data", evt))
	return nil
}

func consumeOrderCanceled(ctx context.Context, broker []string, wg *sync.WaitGroup) {
	defer wg.Done()
	consumer := kafka.NewConsumer(broker, kafka.TopicOrderCanceled, "order-canceled-group")
	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()
		if err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
	}(consumer)

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return handleOrderCanceled(msg)
	})
	if err != nil {
		logger.L().Error("消费 order.canceled 失败", zap.Error(err))
	}
}

func handleOrderCanceled(msg kafkago.Message) error {
	var evt event.OrderCanceledEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal order canceled event failed: %w", err)
	}

	logger.L().Info("收到取消订单事件",
		zap.Int64("order_id", evt.OrderID),
		zap.Int64("product_id", evt.ProductID),
		zap.Int64("count", evt.Count),
		zap.String("reason", evt.Reason),
	)

	return restoreStockByEvent(context.Background(), &evt)
}

func restoreStockByEvent(ctx context.Context, evt *event.OrderCanceledEvent) error {

	productRepo := mysql.NewProductRepository(mySQL)

	// 这里直接调用你现有 repo 的恢复库存方法
	return productRepo.RestoreStock(ctx, evt.ProductID, evt.Count)
}
