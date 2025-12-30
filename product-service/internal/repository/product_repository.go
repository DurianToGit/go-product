package repository

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, id int64) (*domain.Product, error)
	GetWithCreator(ctx context.Context, id int64) (*domain.Product, *domain.User, error)
	List(ctx context.Context, q *dto.ProductQuery) ([]*domain.Product, int64, error)
	DeductStock(ctx context.Context, productID int64, count int64) error
	DeductStockOptimistic(ctx context.Context, productID int64, count int64) (bool, error)
}
