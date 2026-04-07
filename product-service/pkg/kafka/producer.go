package kafka

import (
	"context"
	"fmt"
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

func PublishRaw(ctx context.Context, topic string, data []byte) error {
	if Client == nil {
		return fmt.Errorf("kafka client is nil")
	}

	return Client.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Value: data,
	})
}
