package productclient

import (
	"context"
	productService "product-service/services/product/service"
)

type LocalClient struct {
	svc *productService.ProductService
}

func NewLocalClient(svc *productService.ProductService) Client {
	return &LocalClient{svc: svc}
}

func (c *LocalClient) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	p, err := c.svc.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:    p.ID,
		Title: p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}, nil
}

func (c *LocalClient) GetStock(ctx context.Context, productID int64) (int64, error) {
	return c.svc.GetStock(ctx, productID)
}
