package service

import (
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/repository/mysql/model"
	"product-service/pkg/db"
)

func Register(username, password string) error {
	user := model.UserModel{
		Username: username,
		Password: password,
	}
	return user.Create()
}

func Login(username, password string) (*model.UserModel, error) {
	var user model.UserModel
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	if !user.CheckPassword(password) {
		return nil, errno.ErrPasswordIncorrect
	}
	return &user, nil
}

func GetUserInfo(userId int64) (*dto.UserInfo, error) {
	var user model.UserModel
	err := db.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

func UpdatePassword(userId int64, oldPassword, newPassword string) error {
	var user model.UserModel
	err := db.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		return err
	}
	if !user.CheckPassword(oldPassword) {
		return errno.OldPasswordIncorrect
	}
	return user.UpdatePassword(newPassword)
}
