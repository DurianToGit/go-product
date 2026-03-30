package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"product-service/internal/bootstrap"
	"product-service/pkg/kafka"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserID    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

const groupName = "demo-group-id-2"

func main() {
	cfg := bootstrap.BaseInit()
	reader := kafka.NewConsumer(cfg.Kafka.Addr, "demo.producer", groupName)

	defer func(reader *kafka.Consumer) {
		err := reader.Close()
		if err != nil {
			log.Printf("close error: %v", err)
		}
	}(reader)
	log.Println("starting to read messages")

	ctx := context.Background()
	err := reader.Consume(ctx, handle)
	if err != nil {
		log.Fatalf("consume error: %v", err)
		return
	}
}

func handle(ctx context.Context, msg kafkago.Message) error {
	var d StockDeductRequested
	err2 := json.Unmarshal(msg.Value, &d)
	if err2 != nil {
		log.Fatalf("json error: %v", err2)
	}
	fmt.Printf("received: key=%s value=%v topic=%s group=%s\n",
		string(msg.Key),
		d,
		msg.Topic,
		groupName,
	)
	return nil
}
