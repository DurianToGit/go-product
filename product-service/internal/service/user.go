package service

import (
	"product-service/internal/errno"
	"product-service/internal/model"
	"product-service/pkg/db"
)

func Register(username, password string) error {
	user := model.User{
		Username: username,
		Password: password,
	}
	return user.Create()
}

func Login(username, password string) (*model.User, error) {
	var user model.User
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	if !user.CheckPassword(password) {
		return nil, errno.ErrPasswordIncorrect
	}
	return &user, nil
}
