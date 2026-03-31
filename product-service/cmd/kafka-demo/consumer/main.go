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
		return processMessage(ctx, msg, producerNormal, producerDLQ)
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

func publishToRetry(
	ctx context.Context,
	msg kafkago.Message,
	nextRetry int,
	producer *kafka.Producer,
) error {
	var req StockDeductRequested
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return err
	}
	req.RetryCount = nextRetry

	val, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return producer.Publish(ctx, string(msg.Key), val)
}

func publishToDLQ(
	ctx context.Context,
	msg kafkago.Message,
	retryCount int,
	producer *kafka.Producer,
) error {
	return producer.Publish(ctx, string(msg.Key), msg.Value)
}

func handleStockDeduct(ctx context.Context, msg kafkago.Message) error {
	var req StockDeductRequested
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		// 消息格式坏了，这种重试也没用
		return kafka.NewNonRetryable(fmt.Errorf("invalid message body: %w", err))
	}

	// 业务模拟：商品暂时不可用
	if !isProductIDAvailable(req.ProductID) {
		// 这种以后可能恢复，所以可重试
		return kafka.NewRetryable(
			fmt.Errorf("product %d temporarily unavailable", req.ProductID),
		)
	}

	// 正常处理
	logger.L().Info("处理成功",
		zap.Int64("product_id", req.ProductID),
		zap.Int64("user_id", req.UserID),
	)
	return nil
}

func extractRetryCount(msg kafkago.Message) int {
	retryCount := 0
	var req StockDeductRequested
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		logger.L().Error("解码 Kafka 消息失败，返回超过最大重试次数，转去死信队列", zap.Error(err))
		return maxRetries + 1
	}
	retryCount = req.RetryCount
	return retryCount
}

func processMessage(
	ctx context.Context,
	msg kafkago.Message,
	producerNormal *kafka.Producer,
	producerDLQ *kafka.Producer,
) error {
	err := handleStockDeduct(ctx, msg)
	if err == nil {
		// 成功：提交 offset
		return nil
	}

	retryCount := extractRetryCount(msg)

	// 1. 不可重试错误：直接进 DLQ
	if kafka.IsNonRetryable(err) {
		logger.L().Warn("不可重试错误，发送至 DLQ",
			zap.Error(err),
			zap.Int("retry_count", retryCount),
		)
		return publishToDLQ(ctx, msg, retryCount, producerDLQ)
	}

	// 2. 可重试错误：先看是否超过次数
	if kafka.IsRetryable(err) {
		if retryCount >= maxRetries {
			logger.L().Warn("重试次数已达上限，发送至 DLQ",
				zap.Error(err),
				zap.Int("retry_count", retryCount),
			)
			return publishToDLQ(ctx, msg, retryCount, producerDLQ)
		}

		logger.L().Info("可重试消息，重新发布",
			zap.Error(err),
			zap.Int("retry_count", retryCount),
		)
		return publishToRetry(ctx, msg, retryCount+1, producerNormal)
	}

	// 3. 未分类错误：当前阶段建议按“可重试”处理，或者记 error 后返回
	// 为了 D43 更稳，我建议未分类先按 retryable 处理，但也要受 maxRetries 限制
	if retryCount >= maxRetries {
		logger.L().Error("未知错误已超过最大重试次数，将转至死信队列。",
			zap.Error(err),
			zap.Int("retry_count", retryCount),
		)
		return publishToDLQ(ctx, msg, retryCount, producerDLQ)
	}

	logger.L().Error("未知错误，转去重试",
		zap.Error(err),
		zap.Int("retry_count", retryCount),
	)
	return publishToRetry(ctx, msg, retryCount+1, producerNormal)
}

func isProductIDAvailable(productID int64) bool {
	return productID%3 != 0
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
