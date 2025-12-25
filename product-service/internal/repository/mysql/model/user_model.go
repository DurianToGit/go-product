package model

import (
	"product-service/pkg/db"
	"product-service/pkg/utils"
	"time"
)

type UserModel struct {
	ID        int64  `gorm:"primaryKey"`
	Username  string `gorm:"unique"`
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserModel) TableName() string {
	return "user"
}

func (u *UserModel) Create() error {
	u.Password = utils.HashPassword(u.Password)
	return db.DB.Create(&u).Error
}

func (u *UserModel) CheckPassword(password string) bool {
	return utils.VerifyPassword(password, u.Password)
}

func (u *UserModel) UpdatePassword(password string) error {
	u.Password = utils.HashPassword(password)
	return db.DB.Model(&u).Update("password", u.Password).Error
}
