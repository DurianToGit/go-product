package repository

import (
	"context"
	"product-service/services/user/domain"
	"product-service/services/user/dto"
	"product-service/services/user/repository/mysql/model"
)

type UserRepository interface {
	Register(ctx context.Context, p *model.UserModel) (*domain.User, error)
	Login(ctx context.Context, username, password string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByUserId(ctx context.Context, userId int64) (*domain.User, error)
	UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error
	List(ctx context.Context, query *dto.UserQuery) ([]*domain.User, int64, error)
	Update(ctx context.Context, userId int64, user *dto.UserUpdate) error
}
