package model

import "time"

type EventConsumeLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	EventID       string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_event_consumer"`
	ConsumerGroup string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_event_consumer"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

func (EventConsumeLog) TableName() string {
	return "event_consume_log"
}
