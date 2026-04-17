package productclient

import (
	"context"
	productService "product-service/services/product/service"
)

type LocalClient struct {
	svc *productService.ProductService
}

func (c *LocalClient) ConsumeStockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, eventType string) error {
	return c.svc.ConsumeStockDeductEvent(ctx, eventSource, eventID, productID, count, eventType)
}

func (c *LocalClient) ConsumeRestockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, eventType string) error {
	return c.svc.ConsumeRestockDeductEvent(ctx, eventSource, eventID, productID, count, eventType)
}

func (c *LocalClient) RestoreStock(ctx context.Context, productID int64, count int64, source string) error {
	return c.svc.RestoreStock(ctx, productID, count, source)
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
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}, nil
}

func (c *LocalClient) GetStock(ctx context.Context, productID int64) (int64, error) {
	return c.svc.GetStock(ctx, productID)
}
