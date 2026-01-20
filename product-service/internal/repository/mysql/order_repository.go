package mysql

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"log"
	"product-service/internal/domain"
	"product-service/internal/errno"
	"product-service/internal/repository/mysql/model"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	m := model.OrderModel{
		OrderNo:   order.OrderNo,
		UserID:    order.UserID,
		ProductID: order.ProductID,
		Count:     order.Count,
		Status:    order.Status,
		IdemKey:   order.IdemKey,
	}

	err := r.db.WithContext(ctx).Create(&m).Error
	if err != nil {
		return nil, err
	}
	log.Println("创建的订单ID:", m.ID)
	return m.ToOrderDomain(), err
}

func (r *OrderRepository) Get(ctx context.Context, id int64) (*domain.Order, error) {
	var orderModel model.OrderModel
	err := r.db.WithContext(ctx).First(&orderModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDataNotFound
		}
		return nil, err
	}
	return orderModel.ToOrderDomain(), nil
}

func (r *OrderRepository) GetByUserIdemKey(ctx context.Context, userId int64, idemKey string) (*domain.Order, error) {
	var orderModel model.OrderModel
	err := r.db.WithContext(ctx).Where("user_id = ? and idem_key = ?", userId, idemKey).First(&orderModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDataNotFound
		}
		return nil, err
	}
	return orderModel.ToOrderDomain(), nil
}

func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.OrderModel{}, id).Error
}
