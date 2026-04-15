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
}
