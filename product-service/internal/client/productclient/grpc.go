package productclient

import (
	"context"
	"product-service/internal/errno"
	"time"

	"google.golang.org/grpc"
	"product-service/pkg/pb/productpb"
)

type GRPCClient struct {
	client productpb.ProductServiceClient
}

func NewGRPCClient(conn *grpc.ClientConn) Client {
	return &GRPCClient{
		client: productpb.NewProductServiceClient(conn),
	}
}

func (c *GRPCClient) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.GetProduct(ctx, &productpb.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:    resp.Id,
		Name:  resp.Name,
		Price: resp.Price,
		Stock: resp.Stock,
	}, nil
}

func (c *GRPCClient) GetStock(ctx context.Context, productID int64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.GetStock(ctx, &productpb.GetStockRequest{
		ProductId: productID,
	})
	if err != nil {
		return 0, err
	}
	return resp.Stock, nil
}

func (r *GRPCClient) RestoreStock(ctx context.Context, productID int64, count int64, source string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := r.client.RestoreStock(ctx, &productpb.RestoreStockRequest{
		ProductId: productID,
		Count:     count,
		Source:    source,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return errno.ProductErrRestoreStockFailed
	}

	return nil
}

func (c *GRPCClient) ConsumeStockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, event_type string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	resp, err := c.client.ConsumeStockDeductEvent(ctx, &productpb.ConsumeStockDeductEventRequest{
		EventSource: eventSource,
		EventId:     eventID,
		ProductId:   productID,
		Count:       count,
		EventType:   event_type,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return errno.ProductErrConsumeStockDeductEventFailed
	}
	return nil
}

func (c *GRPCClient) ConsumeRestockDeductEvent(ctx context.Context, eventSource, eventID string, productID int64, count int64, event_type string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	resp, err := c.client.ConsumeRestockDeductEvent(ctx, &productpb.ConsumeStockDeductEventRequest{
		EventSource: eventSource,
		EventId:     eventID,
		ProductId:   productID,
		Count:       count,
		EventType:   event_type,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return errno.ProductErrConsumeRestockDeductEventFailed
	}
	return nil
}
