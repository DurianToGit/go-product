package repository

import (
	"context"
	"product-service/services/product/domain"
	"product-service/services/product/dto"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, q *dto.ProductQuery) ([]*domain.Product, int64, error)
	DeductStock(ctx context.Context, productID int64, count int64) error
	DeductStockOptimistic(ctx context.Context, productID int64, count int64) (bool, error)
	DeductStockAtomic(ctx context.Context, productID int64, count int64) (bool, error)
	ConsumeStockDeductEvent(ctx context.Context, eventSource, eventID string, productID, count int64, eventType string) error
	ConsumeRestockDeductEvent(ctx context.Context, eventSource, eventID string, productID, count int64, eventType string) error
	RestoreStock(ctx context.Context, productID int64, count int64) error
}
