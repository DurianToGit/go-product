package kafka

import (
	"context"
	"encoding/json"
	"product-service/pkg/event"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

var Client *kafkago.Writer

func InitClient(brokers []string) {
	p := &Producer{
		writer: &kafkago.Writer{
			Addr: kafkago.TCP(brokers...),
		},
	}
	Client = p.writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:  kafkago.TCP(brokers...),
			Topic: topic,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func PublishOrderPaid(ctx context.Context, evt event.OrderPaidEvent) error {
	data, _ := json.Marshal(evt)

	return Client.WriteMessages(ctx, kafkago.Message{
		Topic: TopicOrderPaid,
		Value: data,
	})
}
