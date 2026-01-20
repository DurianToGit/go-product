package repository

import (
	"context"
	"product-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	Get(ctx context.Context, id int64) (*domain.Order, error)
	GetByUserIdemKey(ctx context.Context, userId int64, idemKey string) (*domain.Order, error)
	Delete(ctx context.Context, id int64) error
}
