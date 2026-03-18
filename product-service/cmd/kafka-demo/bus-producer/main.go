package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"log"
	"product-service/pkg/eventbus"
	"product-service/pkg/logger"
	"time"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserID    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

func main() {
	bus := eventbus.NewKafkaBus([]string{"localhost:9092"})
	ctx := context.Background()

	defer func(bus *eventbus.KafkaBus) {
		err := bus.Close()
		if err != nil {
			logger.L().Error("关闭kafka失败", zap.Error(err))
		}
	}(bus)

	for i := 0; i < 10; i++ {
		d := StockDeductRequested{
			ProductID: int64(i),
			UserID:    int64(i * 2),
			Count:     int64(i * 3),
		}
		// d转为byte
		bytes, err := json.Marshal(d)
		if err != nil {
			log.Fatalf("json error: %v", err)
		}
		key := fmt.Sprintf("product-%d", i%2)
		err = bus.Publish(ctx, "stock.deduct.requested", key, bytes)
		if err != nil {
			log.Fatalf("write error: %v", err)
		}
		logger.L().Info("发送消息成功", zap.String("key", key), zap.String("value", string(bytes)))
		time.Sleep(time.Second)
	}
}
