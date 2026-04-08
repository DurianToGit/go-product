package mysql

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"product-service/services/order/repository/mysql/model"
)

type EventConsumeLogRepository struct {
	db *gorm.DB
}

func NewEventConsumeLogRepository(db *gorm.DB) *EventConsumeLogRepository {
	return &EventConsumeLogRepository{db: db}
}

// TryConsume
// true  = 首次消费，可以继续执行业务
// false = 已经消费过，直接跳过
func (r *EventConsumeLogRepository) TryConsume(
	ctx context.Context,
	eventID string,
	consumerGroup string,
) (bool, error) {
	log := &model.EventConsumeLog{
		EventID:       eventID,
		ConsumerGroup: consumerGroup,
	}

	err := r.db.WithContext(ctx).Create(log).Error
	if err != nil {
		if isDuplicateErr(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "duplicated") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "uk_event_consumer")
}
