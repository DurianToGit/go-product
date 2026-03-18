package eventbus

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaBus struct {
	writer *kafka.Writer
}

func NewKafkaBus(brokers []string) *KafkaBus {
	return &KafkaBus{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
		},
	}
}

func (b *KafkaBus) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	return b.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})
}

func (b *KafkaBus) Close() error {
	return b.writer.Close()
}
