package mysql

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"product-service/internal/errno"
	"product-service/pkg/logger"
	"product-service/services/order/domain"
	"product-service/services/order/repository/mysql/model"
	"time"
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
		Amount:    order.Amount,
		Status:    order.Status,
		IdemKey:   order.IdemKey,
	}

	err := r.db.WithContext(ctx).Create(&m).Error
	if err != nil {
		return nil, err
	}
	logger.L().Info("创建的订单ID", zap.Any("id", m.ID))
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

func (r *OrderRepository) MarkCancelled(ctx context.Context, id int64) error {
	tx := r.db.WithContext(ctx).
		Model(&model.OrderModel{}).
		Where("id = ?", id).
		Where("status = ?", domain.OrderStatusCreated).
		Update("status", domain.OrderStatusCanceled)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errno.OrderStatusInvalid
	}
	return nil
}

func (r *OrderRepository) CancelExpired(ctx context.Context, deadline time.Time) (int64, []*model.OrderModel, error) {
	var data []*model.OrderModel
	r.db.WithContext(ctx).Model(&model.OrderModel{}).
		Where("status = ? and updated_at < ?", domain.OrderStatusCreated, deadline).
		Find(&data)
	var ids []int64
	for _, v := range data {
		ids = append(ids, v.ID)
		logger.L().Info("取消的订单ID", zap.Any("id", v.ID))
	}
	tx2 := r.db.WithContext(ctx).Model(&model.OrderModel{}).
		Where("id in ?", ids).
		Update("status", domain.OrderStatusCanceled)

	return tx2.RowsAffected, data, tx2.Error
}

func (r *OrderRepository) MarkPaid(ctx context.Context, id int64) error {
	tx := r.db.WithContext(ctx).Model(&model.OrderModel{}).
		Where("id = ?", id).
		Where("status = ?", domain.OrderStatusCreated).
		Update("status", domain.OrderStatusPaid)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errno.OrderStatusInvalid
	}
	return nil
}

func (r *OrderRepository) GetTx(ctx context.Context, tx *gorm.DB, id int64) (*domain.Order, error) {
	var m model.OrderModel
	err := tx.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Order{
		ID:        m.ID,
		OrderNo:   m.OrderNo,
		UserID:    m.UserID,
		ProductID: m.ProductID,
		Count:     m.Count,
		Status:    m.Status,
		IdemKey:   m.IdemKey,
		Amount:    m.Amount,
	}, nil
}

func (r *OrderRepository) MarkPaidTx(ctx context.Context, tx *gorm.DB, id int64) error {
	res := tx.WithContext(ctx).
		Model(&model.OrderModel{}).
		Where("id = ?", id).
		Where("status = ?", domain.OrderStatusCreated).
		Update("status", domain.OrderStatusPaid)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errno.OrderStatusInvalid
	}
	return nil
}
