package service

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/repository"
	"product-service/internal/repository/mysql/model"
	"product-service/pkg/db"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*domain.User, error) {
	user := model.UserModel{
		Username: username,
		Password: password,
	}
	return s.repo.Register(ctx, &user)
}

func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	return s.repo.Login(ctx, username, password)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *UserService) GetByUserId(ctx context.Context, userId int64) (*domain.User, error) {
	return s.repo.GetByUserId(ctx, userId)
}

func (s *UserService) UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	return s.repo.UpdatePassword(ctx, username, oldPassword, newPassword)
}

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
		return nil, errno.UserErrPasswordIncorrect
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
		return errno.UserOldPasswordIncorrect
	}
	return user.UpdatePassword(newPassword)
}
