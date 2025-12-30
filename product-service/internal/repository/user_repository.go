package repository

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/repository/mysql/model"
)

type UserRepository interface {
	Register(ctx context.Context, p *model.UserModel) (*domain.User, error)
	Login(ctx context.Context, username, password string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByUserId(ctx context.Context, userId int64) (*domain.User, error)
	UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error
}
