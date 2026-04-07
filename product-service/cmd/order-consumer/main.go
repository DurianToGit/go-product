package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/pkg/db"
	"product-service/pkg/event"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/services/product/repository/mysql"
	"sync"
	"syscall"
)

var mySQL *gorm.DB
var productRepo *mysql.ProductRepository

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

	productRepo = mysql.NewProductRepository(mySQL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go consumeOrderCanceled(ctx, cfg.Kafka.Addr, &wg)
	go consumeOrderPaid(ctx, cfg.Kafka.Addr, &wg)

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
		return handleOrderPaid(ctx, msg)
	})
	if err != nil {
		logger.L().Error("消费 order.paid 失败", zap.Error(err))
	}
}

func handleOrderPaid(ctx context.Context, msg kafkago.Message) error {
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
		return handleOrderCanceled(ctx, msg)
	})
	if err != nil {
		logger.L().Error("消费 order.canceled 失败", zap.Error(err))
	}
}

func handleOrderCanceled(ctx context.Context, msg kafkago.Message) error {
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
	return productRepo.RestoreStock(ctx, evt.ProductID, evt.Count)
}
