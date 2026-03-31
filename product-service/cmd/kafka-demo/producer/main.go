package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"product-service/internal/bootstrap"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"time"
)

type StockDeductRequested struct {
	ProductID  int64 `json:"product_id"`
	UserID     int64 `json:"user_id"`
	Count      int64 `json:"count"`
	RetryCount int   `json:"retry_count"`
}

func main() {
	cfg := bootstrap.BaseInit()
	writer := kafka.NewProducer(cfg.Kafka.Addr, "demo.stock.deduct.requested")

	defer func(writer *kafka.Producer) {
		err := writer.Close()
		if err != nil {
			logger.L().Error("close error", zap.Error(err))
		}
	}(writer)

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		d := StockDeductRequested{
			ProductID: int64(i),
			UserID:    int64(i * 2),
			Count:     int64(i * 3),
		}
		// d转为byte
		bytes, err := json.Marshal(d)
		if err != nil {
			logger.L().Error("json error", zap.Error(err))
		}
		key := fmt.Sprintf("product-test-%d", i%2)

		err = writer.Publish(ctx, key, bytes)
		if err != nil {
			logger.L().Error("write error", zap.Error(err))
		}

		logger.L().Info("sent", zap.String("key", key), zap.ByteString("value", bytes))
		time.Sleep(time.Second)
	}
}
