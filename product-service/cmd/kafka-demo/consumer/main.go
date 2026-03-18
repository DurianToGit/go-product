package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

type StockDeductRequested struct {
	ProductID int64 `json:"product_id"`
	UserId    int64 `json:"user_id"`
	Count     int64 `json:"count"`
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "stock.deduct.requested",
		GroupID: "demo-group2",
	})

	defer reader.Close()
	log.Println("starting to read messages")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		var d StockDeductRequested
		err2 := json.Unmarshal(msg.Value, &d)
		if err != nil {
			log.Fatalf("json error: %v", err2)
		}

		fmt.Printf("received: key=%s value=%v partition=%d offset=%d\n",
			string(msg.Key),
			d,
			msg.Partition,
			msg.Offset,
		)
	}
}
