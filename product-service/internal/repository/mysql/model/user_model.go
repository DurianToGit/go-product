package model

import (
	"product-service/pkg/db"
	"product-service/pkg/utils"
)

type UserModel struct {
	BaseModel
	Username string `gorm:"type:varchar(64);not null;uniqueIndex:uk_username"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (UserModel) TableName() string {
	return "users"
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
