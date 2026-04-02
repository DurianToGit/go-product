package mysql

import (
	"context"

	"gorm.io/gorm"
	"product-service/services/order/repository/mysql/model"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) CreateTx(ctx context.Context, tx *gorm.DB, event *model.OutboxEventModel) error {
	return tx.WithContext(ctx).Create(event).Error
}

func (r *OutboxRepository) FindPending(ctx context.Context, limit int) ([]*model.OutboxEventModel, error) {
	var events []*model.OutboxEventModel
	err := r.db.WithContext(ctx).
		Model(&model.OutboxEventModel{}).
		Where("status = ?", 0).
		Order("id ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     1,
			"last_error": "",
		}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id int64, lastErr string) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     2,
			"last_error": lastErr,
		}).Error
}

func (r *OutboxRepository) IncrementRetry(ctx context.Context, id int64, lastErr string) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  lastErr,
		}).Error
}
