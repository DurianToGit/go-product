package model

import (
	"product-service/pkg/db"
	"product-service/pkg/utils"
)

type User struct {
	ID       int64  `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

func (User) TableName() string {
	return "user"
}

func (u *User) Create() error {
	u.Password = utils.HashPassword(u.Password)
	return db.DB.Create(&u).Error
}

func (u *User) CheckPassword(password string) bool {
	return utils.VerifyPassword(password, u.Password)
}
