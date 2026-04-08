package repository

import "context"

type EventConsumeLogRepository interface {
	TryConsume(ctx context.Context, eventID string, consumerGroup string) (bool, error)
}
