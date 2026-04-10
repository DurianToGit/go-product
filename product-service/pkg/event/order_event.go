package event

import (
	"product-service/pkg/pb/orderpb"
)

type OrderPaidEvent struct {
	OrderID int64
	UserID  int64
	Amount  int64
	PaidAt  int64
}

type OrderCanceledEvent struct {
	OrderID    int64  `json:"order_id"`
	UserID     int64  `json:"user_id"`
	ProductID  int64  `json:"product_id"`
	Count      int64  `json:"count"`
	Reason     string `json:"reason"`
	CanceledAt int64  `json:"canceled_at"`
}

func (e *OrderPaidEvent) ToPb() *orderpb.OrderPaidEvent {
	return &orderpb.OrderPaidEvent{
		OrderId: e.OrderID,
		UserId:  e.UserID,
		Amount:  e.Amount,
		PaidAt:  e.PaidAt,
	}
}

func (e *OrderCanceledEvent) ToPb() *orderpb.OrderCanceledEvent {
	return &orderpb.OrderCanceledEvent{
		OrderId:    e.OrderID,
		UserId:     e.UserID,
		ProductId:  e.ProductID,
		Count:      e.Count,
		Reason:     e.Reason,
		CanceledAt: e.CanceledAt,
	}
}
