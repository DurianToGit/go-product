package domain

import "time"

type Order struct {
	ID        int64     `json:"id"`
	OrderNo   string    `json:"order_no"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Count     int       `json:"count"`
	Status    string    `json:"status"`
	IdemKey   string    `json:"idem_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const OrderStatusCreated = "created"
const OrderStatusPaid = "paid"
const OrderStatusCanceled = "canceled"
