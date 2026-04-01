package main

import (
	"context"
	"encoding/json"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"product-service/internal/bootstrap"
	"product-service/pkg/event"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
)

func main() {
	cfg := bootstrap.BaseInit()
	logger.InitFromEnv("order-consumer")
	defer logger.Sync()
	logger.L().Info("启动订单支付消费者")
	reader := kafka.NewConsumer(cfg.Kafka.Addr, kafka.TopicOrderPaid, "order-consumer-group")
	defer func() {
		if err := reader.Close(); err != nil {
			logger.L().Error("关闭消费者失败", zap.Error(err))
		}
	}()
	ctx := context.Background()
	err := reader.Consume(ctx, func(ctx context.Context, msg kafkago.Message) error {
		handleOrderPaid(msg)
		return nil
	})
	if err != nil {
		logger.L().Error("消费订单支付事件失败：", zap.Error(err))
	}
}

func handleOrderPaid(msg kafkago.Message) {
	var evt event.OrderPaidEvent
	err := json.Unmarshal(msg.Value, &evt)
	if err != nil {
		logger.L().Error("解析订单支付事件失败：", zap.Error(err), zap.Any("data", evt))
		return
	}

	logger.L().Info("增加积分：", zap.Int64("user_id", evt.UserID))
}
