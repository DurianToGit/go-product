package domain

import "time"

type Order struct {
	ID        int64     `json:"id"`
	OrderNo   string    `json:"order_no"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Count     int       `json:"count"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	IdemKey   string    `json:"idem_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusCreated  OrderStatus = "created"
	OrderStatusPaid     OrderStatus = "paid"
	OrderStatusCanceled OrderStatus = "canceled"
)

const ProductEventTypeStockDeducted = "stock_deducted"
const ProductEventTypeRestockDeducted = "restock_deducted"

var orderTransitions = map[OrderStatus]map[OrderStatus]struct{}{
	OrderStatusCreated: {
		OrderStatusPaid:     {},
		OrderStatusCanceled: {},
	},
	OrderStatusPaid:     {},
	OrderStatusCanceled: {},
}

func CanTransition(from, to OrderStatus) bool {
	next, ok := orderTransitions[from]
	if !ok {
		return false
	}
	_, ok = next[to]
	return ok
}
