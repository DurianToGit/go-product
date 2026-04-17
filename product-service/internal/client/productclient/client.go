package productclient

import "context"

type Product struct {
	ID    int64
	Name  string
	Price int64
	Stock int64
}

type Client interface {
	GetProduct(ctx context.Context, productID int64) (*Product, error)
	GetStock(ctx context.Context, productID int64) (int64, error)
	RestoreStock(ctx context.Context, productID int64, count int64, source string) error
	ConsumeStockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, eventType string) error
	ConsumeRestockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, eventType string) error
}
