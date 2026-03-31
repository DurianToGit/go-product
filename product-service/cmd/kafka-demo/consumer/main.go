package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"sync"
	"syscall"
)

type StockDeductRequested struct {
	ProductID  int64 `json:"product_id"`
	UserID     int64 `json:"user_id"`
	Count      int64 `json:"count"`
	RetryCount int   `json:"retry_count"`
}

const (
	groupName    = "demo-group-id-2"
	dlqGroupName = "demo-group-id-2-dlq"
	maxRetries   = 3
	topicNormal  = "demo.stock.deduct.requested"
	topicDLQ     = "demo.stock.deduct.requested.dlq"
)

func main() {
	cfg := bootstrap.BaseInit()

	reader := kafka.NewConsumer(cfg.Kafka.Addr, topicNormal, groupName)
	readerDLQ := kafka.NewConsumer(cfg.Kafka.Addr, topicDLQ, dlqGroupName)
	producerNormal := kafka.NewProducer(cfg.Kafka.Addr, topicNormal)
	producerDLQ := kafka.NewProducer(cfg.Kafka.Addr, topicDLQ)

	logger.L().Info("开始监听消息",
		zap.String("normal_group", groupName),
		zap.String("dlq_group", dlqGroupName))

	defer func() {
		if err := reader.Close(); err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
		if err := readerDLQ.Close(); err != nil {
			logger.L().Error("关闭死信队列消费者失败", zap.Error(err))
		}
		if err := producerNormal.Close(); err != nil {
			logger.L().Error("关闭生产者失败", zap.Error(err))
		}
		if err := producerDLQ.Close(); err != nil {
			logger.L().Error("关闭死信队列生产者失败", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		consumeNormal(ctx, reader, producerNormal, producerDLQ)
	}()

	go func() {
		defer wg.Done()
		consumeDLQ(ctx, readerDLQ)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.L().Info("收到退出信号，开始优雅关闭...")
	cancel()

	wg.Wait()
	logger.L().Info("所有消费者已停止")
}

func consumeNormal(ctx context.Context, reader *kafka.Consumer, producerNormal, producerDLQ *kafka.Producer) {
	logger.L().Info("普通队列消费者已启动")
	err := reader.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		var d StockDeductRequested
		if err := json.Unmarshal(msg.Value, &d); err != nil {
			logger.L().Error("unmarshal kafka message failed", zap.Error(err))
			return fmt.Errorf("unmarshal kafka message failed: %w", err)
		}

		if !isProductIDAvailable(d.ProductID) {
			return handleProductUnavailable(ctx, producerNormal, producerDLQ, msg.Key, d)
		}

		logger.L().Info("收到消息",
			zap.String("key", string(msg.Key)),
			zap.Any("value", d),
			zap.String("topic", msg.Topic),
			zap.String("group", groupName),
		)
		return nil
	})
	if err != nil {
		logger.L().Error("消费队列异常", zap.Error(err))
		return
	}
}

func consumeDLQ(ctx context.Context, reader *kafka.Consumer) {
	logger.L().Info("死信队列消费者已启动")
	err := reader.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		var d StockDeductRequested
		if err := json.Unmarshal(msg.Value, &d); err != nil {
			logger.L().Error("解码 Kafka 消息失败", zap.Error(err))
			return fmt.Errorf("解码 Kafka 消息失败: %w", err)
		}
		logDeadLetterMessage(d)
		return nil
	})
	if err != nil {
		logger.L().Error("死信队列消费者异常", zap.Error(err))
	}
}

func isProductIDAvailable(productID int64) bool {
	return productID%3 != 0
}

func handleProductUnavailable(
	ctx context.Context,
	producerNormal, producerDLQ *kafka.Producer,
	key []byte,
	data StockDeductRequested,
) error {
	data.RetryCount++

	val, errMarshal := json.Marshal(data)
	if errMarshal != nil {
		logger.L().Error("marshal kafka message failed", zap.Error(errMarshal))
		return errMarshal
	}

	if data.RetryCount > maxRetries {
		if err := producerDLQ.Publish(ctx, string(key), val); err != nil {
			logger.L().Error("发布消息到死信队列失败",
				zap.String("key", string(key)),
				zap.Error(err))
			return err
		}
		logger.L().Warn("消息进入死信队列",
			zap.String("key", string(key)),
			zap.Int("retry_count", data.RetryCount),
			zap.Int64("product_id", data.ProductID))
		return nil
	}

	if err := producerNormal.Publish(ctx, string(key), val); err != nil {
		logger.L().Error("重新发布消息失败",
			zap.String("key", string(key)),
			zap.Error(err))
		return err
	}

	logger.L().Info("消息已重新投递",
		zap.String("key", string(key)),
		zap.Int("retry_count", data.RetryCount),
		zap.Int64("product_id", data.ProductID))
	return nil
}

func logDeadLetterMessage(data StockDeductRequested) {
	logger.L().Error("=== 死信消息详情 ===",
		zap.Int64("product_id", data.ProductID),
		zap.Int64("user_id", data.UserID),
		zap.Int64("count", data.Count),
		zap.Int("retry_count", data.RetryCount),
		zap.String("reason", "产品 ID 不可用且超过最大重试次数"),
	)
}
