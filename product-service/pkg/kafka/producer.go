package kafka

import "context"

type EventBus interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}