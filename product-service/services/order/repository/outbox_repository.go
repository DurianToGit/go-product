package repository

import (
	"context"
	"gorm.io/gorm"
	"product-service/services/order/repository/mysql/model"
)

type OutboxRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, event *model.OutboxEventModel) error
	FindPending(ctx context.Context, limit int) ([]*model.OutboxEventModel, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, lastErr string) error
	IncrementRetry(ctx context.Context, id int64, lastErr string) error
}
