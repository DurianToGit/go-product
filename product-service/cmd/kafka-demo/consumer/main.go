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
	"time"
)

type StockDeductRequested struct {
	ProductID  int64 `json:"product_id"`
	UserID     int64 `json:"user_id"`
	Count      int64 `json:"count"`
	RetryCount int   `json:"retry_count"`
}

const (
	groupName    = "demo-group-id-2"
	groupName1s  = "demo-group-id-2-1s"
	groupName5s  = "demo-group-id-2-5s"
	groupName30s = "demo-group-id-2-30s"
	GroupNameDlq = "demo-group-id-2-dlq"
	topicNormal  = "demo.stock.deduct.requested"
	topic1s      = "demo.stock.deduct.requested.1s"
	topic5s      = "demo.stock.deduct.requested.5s"
	topic30s     = "demo.stock.deduct.requested.30s"
	topicDLQ     = "demo.stock.deduct.requested.dlq"
)

func main() {
	cfg := bootstrap.BaseInit()

	reader := kafka.NewConsumer(cfg.Kafka.Addr, topicNormal, groupName)
	reader1s := kafka.NewConsumer(cfg.Kafka.Addr, topic1s, groupName1s)
	reader5s := kafka.NewConsumer(cfg.Kafka.Addr, topic5s, groupName5s)
	reader30s := kafka.NewConsumer(cfg.Kafka.Addr, topic30s, groupName30s)
	readerDLQ := kafka.NewConsumer(cfg.Kafka.Addr, topicDLQ, GroupNameDlq)
	producerNormal := kafka.NewProducer(cfg.Kafka.Addr, topicNormal)
	producer1s := kafka.NewProducer(cfg.Kafka.Addr, topic1s)
	producer5s := kafka.NewProducer(cfg.Kafka.Addr, topic5s)
	producer30s := kafka.NewProducer(cfg.Kafka.Addr, topic30s)
	producerDLQ := kafka.NewProducer(cfg.Kafka.Addr, topicDLQ)

	logger.L().Info("开始监听消息",
		zap.String("normal_group", groupName),
		zap.String("dlq_group", GroupNameDlq))

	defer func() {
		if err := reader.Close(); err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
		if err := reader1s.Close(); err != nil {
			logger.L().Error("关闭1s队列消费者失败", zap.Error(err))
		}
		if err := reader5s.Close(); err != nil {
			logger.L().Error("关闭5s队列消费者失败", zap.Error(err))
		}
		if err := reader30s.Close(); err != nil {
			logger.L().Error("关闭30s队列消费者失败", zap.Error(err))
		}
		if err := readerDLQ.Close(); err != nil {
			logger.L().Error("关闭死信队列消费者失败", zap.Error(err))
		}
		if err := producerNormal.Close(); err != nil {
			logger.L().Error("关闭生产者失败", zap.Error(err))
		}
		if err := producer1s.Close(); err != nil {
			logger.L().Error("关闭1s队列生产者失败", zap.Error(err))
		}
		if err := producer5s.Close(); err != nil {
			logger.L().Error("关闭5s队列生产者失败", zap.Error(err))
		}
		if err := producer30s.Close(); err != nil {
			logger.L().Error("关闭30s队列生产者失败", zap.Error(err))
		}
		if err := producerDLQ.Close(); err != nil {
			logger.L().Error("关闭死信队列生产者失败", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(5)

	go func() {
		defer wg.Done()
		consumeNormal(ctx, reader, producer1s, producer5s, producer30s, producerDLQ)
	}()

	go func() {
		defer wg.Done()
		consumeRetryTopic(ctx, reader1s, 1*time.Second, producerNormal, "retry-1s")
	}()

	go func() {
		defer wg.Done()
		consumeRetryTopic(ctx, reader5s, 5*time.Second, producerNormal, "retry-5s")
	}()

	go func() {
		defer wg.Done()
		consumeRetryTopic(ctx, reader30s, 30*time.Second, producerNormal, "retry-30s")
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

func consumeNormal(ctx context.Context, reader *kafka.Consumer, producer1s, producer5s, producer30s, producerDLQ *kafka.Producer) {
	logger.L().Info("普通队列消费者已启动")
	err := reader.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return processMessage(ctx, msg, producer1s, producer5s, producer30s, producerDLQ)
	})
	if err != nil {
		logger.L().Error("消费队列异常", zap.Error(err))
		return
	}
}

func consumeRetryTopic(
	ctx context.Context,
	reader *kafka.Consumer,
	delay time.Duration,
	producerNormal *kafka.Producer,
	name string,
) {
	logger.L().Info("retry consumer started",
		zap.String("name", name),
		zap.Duration("delay", delay),
	)

	err := reader.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		logger.L().Info("retry message received",
			zap.String("name", name),
			zap.String("key", string(msg.Key)),
			zap.Duration("delay", delay),
		)

		time.Sleep(delay)

		logger.L().Info("retry message republish to main topic",
			zap.String("name", name),
			zap.String("key", string(msg.Key)),
		)

		return producerNormal.Publish(ctx, string(msg.Key), msg.Value)
	})

	if err != nil {
		logger.L().Error("retry consumer failed",
			zap.String("name", name),
			zap.Error(err),
		)
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

func extractRetryCount(msg kafkago.Message) (int, error) {
	var req StockDeductRequested
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return 0, err
	}
	return req.RetryCount, nil
}

func processMessage(
	ctx context.Context,
	msg kafkago.Message,
	producerRetry1s *kafka.Producer,
	producerRetry5s *kafka.Producer,
	producerRetry30 *kafka.Producer,
	producerDLQ *kafka.Producer,
) error {
	err := handleStockDeduct(ctx, msg)
	if err == nil {
		return nil
	}

	retryCount, err2 := extractRetryCount(msg)
	if err2 != nil {
		logger.L().Warn("extract retry count failed, send to dlq",
			zap.Error(err2),
		)
		return producerDLQ.Publish(ctx, string(msg.Key), msg.Value)
	}

	// 不可重试：直接 DLQ
	if kafka.IsNonRetryable(err) {
		logger.L().Warn("non-retryable error, send to dlq",
			zap.Error(err),
			zap.Int("retry_count", retryCount),
		)
		return producerDLQ.Publish(ctx, string(msg.Key), msg.Value)
	}

	// 可重试 / 未分类错误：进入下一层延迟 topic
	nextCount := retryCount + 1
	targetTopic := nextRetryTopic(nextCount)

	if targetTopic == topicDLQ {
		logger.L().Warn("retry exceeded, send to dlq",
			zap.Error(err),
			zap.Int("retry_count", retryCount),
		)
		return producerDLQ.Publish(ctx, string(msg.Key), msg.Value)
	}

	val, err3 := rebuildMessageWithRetryCount(msg, nextCount)
	if err3 != nil {
		return err3
	}

	logger.L().Info("send to retry topic",
		zap.Error(err),
		zap.Int("retry_count", nextCount),
		zap.String("target_topic", targetTopic),
	)

	switch targetTopic {
	case topic1s:
		return producerRetry1s.Publish(ctx, string(msg.Key), val)
	case topic5s:
		return producerRetry5s.Publish(ctx, string(msg.Key), val)
	case topic30s:
		return producerRetry30.Publish(ctx, string(msg.Key), val)
	default:
		return producerDLQ.Publish(ctx, string(msg.Key), msg.Value)
	}
}

func rebuildMessageWithRetryCount(msg kafkago.Message, retryCount int) ([]byte, error) {
	var req StockDeductRequested
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return nil, err
	}
	req.RetryCount = retryCount
	return json.Marshal(req)
}

func nextRetryTopic(retryCount int) string {
	switch retryCount {
	case 1:
		return topic1s
	case 2:
		return topic5s
	case 3:
		return topic30s
	default:
		return topicDLQ
	}
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
