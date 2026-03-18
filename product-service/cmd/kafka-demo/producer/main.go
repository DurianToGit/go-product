package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserId    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

func main() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "stock.deduct.requested",
	})

	defer writer.Close()

	for i := 0; i < 10; i++ {
		d := StockDeductRequested{
			ProductID: int64(i),
			UserId:    int64(i * 2),
			Count:     int64(i * 3),
		}
		// d转为byte
		bytes, _ := json.Marshal(d)
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("product-%d", i%2)), // 故意让 key 重复
			Value: bytes,
		}

		err := writer.WriteMessages(context.Background(), msg)
		if err != nil {
			log.Fatalf("write error: %v", err)
		}

		fmt.Println("sent:", string(msg.Key), string(msg.Value))
		time.Sleep(time.Second)
	}
}
