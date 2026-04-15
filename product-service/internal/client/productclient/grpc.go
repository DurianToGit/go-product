package productclient

import (
	"context"
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
		Title: resp.Name,
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
