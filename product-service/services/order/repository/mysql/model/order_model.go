package model

import (
	"product-service/services/order/domain"
	"time"
)

type OrderModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
	OrderNo   string    `gorm:"not null;type:varchar(100);uniqueIndex:idx_order_no"`
	UserID    int64     `gorm:"not null;index:idx_user_id;uniqueIndex:uk_user_idem,priority:1"`
	ProductID int64     `gorm:"not null;index:idx_product_id"`
	Count     int       `gorm:"not null;default:1"`
	Status    string    `gorm:"type:varchar(32);not null;default:'created'"`
	IdemKey   string    `gorm:"type:varchar(128);not null;uniqueIndex:uk_user_idem,priority:2"`
	// user_id + idem_key 复合唯一, IdemKey（幂等键）是：客户端为“一次业务意图”生成的唯一标识，用来保证“同一意图只创建一笔订单”。
}

func (OrderModel) TableName() string {
	return "orders"
}

func (order *OrderModel) ToOrderDomain() *domain.Order {
	return &domain.Order{
		ID:        order.ID,
		OrderNo:   order.OrderNo,
		UserID:    order.UserID,
		ProductID: order.ProductID,
		Count:     order.Count,
		Status:    order.Status,
		IdemKey:   order.IdemKey,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}
