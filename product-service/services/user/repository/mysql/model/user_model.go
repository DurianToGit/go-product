package model

import "time"

type UserModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
	Version   int       `gorm:"not null;default:1"`
	Username  string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_username"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Status    int       `gorm:"not null;default:1;index:idx_status"`
}

func (UserModel) TableName() string {
	return "users"
}
