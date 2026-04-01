package event

type OrderPaidEvent struct {
	OrderID int64
	UserID  int64
	Amount  int64
	PaidAt  int64
}
