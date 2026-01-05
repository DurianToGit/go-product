package service

import (
	"context"
	"encoding/json"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/repository"
	"product-service/internal/repository/mysql/model"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"time"
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
	key := rediskeys.UserInfoKey(userId)
	// 从缓存中获取
	if val, err := redis.Client.Get(ctx, key).Result(); err == nil {
		var user domain.User
		if err = json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
		_ = redis.Client.Del(ctx, key).Err()
	}
	// 获取数据
	user, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 缓存
	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	err = redis.Client.Set(ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	return s.repo.UpdatePassword(ctx, username, oldPassword, newPassword)
}

func (s *UserService) List(ctx context.Context, query *dto.UserQuery) ([]*domain.User, int64, error) {
	return s.repo.List(ctx, query)
}

func (s *UserService) Update(ctx context.Context, userId int64, req *dto.UserUpdate) error {
	return s.repo.Update(ctx, userId, req)
}
