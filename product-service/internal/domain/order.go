package domain

type Order struct {
	ID        int64  `json:"id"`
	OrderNo   string `json:"order_no"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Count     int    `json:"count"`
	Status    string `json:"status"`
	IdemKey   string `json:"idem_key"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

const OrderStatusCreated = "created"
const OrderStatusPaid = "paid"
const OrderStatusCanceled = "canceled"
