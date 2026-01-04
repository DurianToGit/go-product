package model

import "time"

type BaseModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Status    int       `gorm:"not null;default:1;index:idx_status"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}
