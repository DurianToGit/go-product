package repository

import (
	"context"
	"product-service/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, page, size int) ([]*domain.Product, error)
	DeductStock(ctx context.Context, productID int64, count int64) error
	DeductStockOptimistic(ctx context.Context, productID int64, count int64) (bool, error)
}
