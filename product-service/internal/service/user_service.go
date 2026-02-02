package service

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/sync/singleflight"
	"math/rand"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/repository"
	"product-service/internal/repository/mysql/model"
	"product-service/pkg/cache"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"time"
)

type UserService struct {
	repo       repository.UserRepository
	localCache *cache.LocalCache
}

var (
	userSF singleflight.Group
)

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo:       repo,
		localCache: cache.NewLocalCache(30 * time.Second),
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
	// L1缓存
	if val, ok := s.localCache.Get(key); ok {
		if string(val) == domain.DataCacheNil {
			return nil, errno.UserErrNotFound
		}
		var user domain.User
		if err := json.Unmarshal(val, &user); err == nil {
			return &user, nil
		}
		s.localCache.Delete(key)
	}
	// singleflight 合并回源
	v, err, _ := userSF.Do(key, func() (any, error) {
		// redis
		var (
			val  []byte
			rerr error
		)
		rerr = redis.Do(ctx, func() error {
			val, rerr = redis.Client.Get(ctx, key).Bytes()
			return rerr
		})
		// val, rerr := redis.Client.Get(ctx, key).Bytes()
		if rerr == nil {
			s.localCache.Set(key, val)
			if string(val) == domain.DataCacheNil {
				return nil, errno.UserErrNotFound
			}
			var user domain.User
			if err := json.Unmarshal(val, &user); err == nil {
				return &user, nil
			}
		}
		// DB
		u, uerr := s.repo.GetByUserId(ctx, userId)
		if uerr != nil {
			if errors.Is(uerr, errno.UserErrNotFound) {
				s.localCache.Set(key, []byte(domain.DataCacheNil))
				_ = redis.Client.Set(ctx, key, domain.DataCacheNil, 30*time.Second).Err()
			}
			return nil, uerr
		}
		data, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}
		s.localCache.Set(key, data)

		ttl := 5*time.Minute + time.Duration(rand.Intn(60))*time.Second
		cerr := redis.Do(ctx, func() error {
			return redis.Client.Set(ctx, key, data, ttl).Err()
		})
		if cerr != nil {
			return nil, cerr
		}
		return u, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(*domain.User), nil
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
