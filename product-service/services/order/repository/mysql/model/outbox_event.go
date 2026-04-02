package model

import "time"

// OutboxEventModel 对应数据库表 outbox_event
type OutboxEventModel struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:主键"`
	EventType  string    `gorm:"type:varchar(50);not null;index:idx_event_type;comment:事件类型"`
	BizID      int64     `gorm:"not null;default:0;index:idx_biz_id"`
	Payload    string    `gorm:"type:text;not null;comment=kafka消息体"`
	Status     uint8     `gorm:"type:tinyint;not null;default:0;comment:状态是否发送"`
	RetryCount int       `gorm:"default:0;comment:重试次数"`
	LastError  string    `gorm:"type:varchar(500);not null;default:''"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (OutboxEventModel) TableName() string {
	return "outbox_event"
}
