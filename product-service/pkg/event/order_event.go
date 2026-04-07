package event

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
