package orderclient

import (
	"context"
)

type orderInfo struct {
	OrderId   int64
	OrderNo   string
	UserId    int64
	ProductId int64
	Count     int64
	Amount    int64
	Status    string
}

type Client interface {
	Create(ctx context.Context, userId int64, productId int64, count int64, idemKey string) (*orderInfo, error)
	Cancel(ctx context.Context, orderId int64, reason string) error
	Get(ctx context.Context, orderId int64) (*orderInfo, error)
	Pay(ctx context.Context, orderId int64) error
}
