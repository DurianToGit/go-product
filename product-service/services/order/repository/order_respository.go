package repository

import (
	"context"
	"product-service/internal/domain"
	"product-service/services/order/repository/mysql/model"
	"time"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	Get(ctx context.Context, id int64) (*domain.Order, error)
	GetByUserIdemKey(ctx context.Context, userId int64, idemKey string) (*domain.Order, error)
	Delete(ctx context.Context, id int64) error
	CancelExpired(ctx context.Context, deadline time.Time) (int64, []*model.OrderModel, error)
}
