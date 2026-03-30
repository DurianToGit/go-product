package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"product-service/internal/bootstrap"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserID    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

const groupName = "demo-group-id-2"

func main() {
	cfg := bootstrap.BaseInit()
	reader := kafka.NewConsumer(cfg.Kafka.Addr, "demo.stock.deduct.requested", groupName)

	defer func(reader *kafka.Consumer) {
		err := reader.Close()
		if err != nil {
			logger.L().Info("close error", zap.Error(err))
		}
	}(reader)
	logger.L().Info("consumer started")

	ctx := context.Background()
	err := reader.Consume(ctx, handle)
	if err != nil {
		logger.L().Error("consume error", zap.Error(err))
		return
	}
}

func handle(ctx context.Context, msg kafkago.Message) error {
	var d StockDeductRequested
	err2 := json.Unmarshal(msg.Value, &d)
	if err2 != nil {
		return fmt.Errorf("unmarshal kafka message failed: %w", err2)
	}
	logger.L().Info("received message",
		zap.String("key", string(msg.Key)),
		zap.Any("value", d),
		zap.String("topic", msg.Topic),
		zap.String("group", groupName),
	)
	return nil
}
