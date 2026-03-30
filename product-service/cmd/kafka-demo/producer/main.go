package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"product-service/internal/bootstrap"
	"product-service/pkg/kafka"
	"time"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserID    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

func main() {
	cfg := bootstrap.BaseInit()
	writer := kafka.NewProducer(cfg.Kafka.Addr, "demo.producer")

	defer func(writer *kafka.Producer) {
		err := writer.Close()
		if err != nil {
			log.Printf("close error: %v", err)
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
			log.Fatalf("json error: %v", err)
		}
		key := fmt.Sprintf("product-test-%d", i%2)

		err = writer.Publish(ctx, key, bytes)
		if err != nil {
			log.Fatalf("write error: %v", err)
		}

		fmt.Println("sent:", key, bytes)
		time.Sleep(time.Second)
	}
}
