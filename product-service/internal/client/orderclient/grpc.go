package orderclient

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"product-service/pkg/pb/orderpb"
)

type GRPCClient struct {
	client orderpb.OrderServiceClient
}

func NewGRPCClient(conn *grpc.ClientConn) Client {
	return &GRPCClient{
		client: orderpb.NewOrderServiceClient(conn),
	}
}

func (c *GRPCClient) Create(ctx context.Context, userId int64, productId int64, count int64, idemKey string) (*orderInfo, error) {
	resp, err := c.client.Create(ctx, &orderpb.CreateRequest{
		UserId:    userId,
		ProductId: productId,
		Count:     count,
		IdemKey:   idemKey,
	})
	if err != nil {
		return nil, err
	}
	return &orderInfo{
		OrderId:   resp.Id,
		OrderNo:   resp.OrderNo,
		UserId:    resp.UserId,
		ProductId: resp.ProductId,
		Count:     resp.Count,
		Amount:    resp.Amount,
		Status:    resp.Status,
	}, nil
}

func (c *GRPCClient) Cancel(ctx context.Context, orderId int64, reason string) error {
	resp, err := c.client.Cancel(ctx, &orderpb.CancelRequest{
		OrderId: orderId,
		Reason:  reason,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("cancel order failed")
	}
	return nil
}

func (c *GRPCClient) Get(ctx context.Context, orderId int64) (*orderInfo, error) {
	resp, err := c.client.Get(ctx, &orderpb.GetRequest{
		OrderId: orderId,
	})
	if err != nil {
		return nil, err
	}
	return &orderInfo{
		OrderId:   resp.Id,
		OrderNo:   resp.OrderNo,
		UserId:    resp.UserId,
		ProductId: resp.ProductId,
		Count:     resp.Count,
		Amount:    resp.Amount,
		Status:    resp.Status,
	}, nil
}

func (c *GRPCClient) Pay(ctx context.Context, orderId int64) error {
	resp, err := c.client.Pay(ctx, &orderpb.PayRequest{
		OrderId: orderId,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("pay order failed")
	}
	return nil
}
