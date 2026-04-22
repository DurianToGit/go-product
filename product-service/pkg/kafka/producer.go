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

// Dial 验证 Kafka broker 是否可达，超时则返回错误
func Dial(ctx context.Context, brokers []string) error {
	cl := &kafkago.Client{Addr: kafkago.TCP(brokers[0])}
	_, err := cl.Metadata(ctx, &kafkago.MetadataRequest{})
	if err != nil {
		return fmt.Errorf("kafka broker %v 不可达: %w", brokers, err)
	}
	return nil
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

func PublishRaw(ctx context.Context, topic string, key []byte, data []byte) error {
	if Client == nil {
		return fmt.Errorf("kafka client is nil")
	}

	return Client.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   key,
		Value: data,
	})
}
