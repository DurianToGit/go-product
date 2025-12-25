package mysql

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"product-service/internal/domain"
	"product-service/internal/errno"
	"product-service/internal/repository/mysql/model"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Product, error) {
	fmt.Println("参数page=", page, "pageSize=", pageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var list []*model.ProductModel
	err := r.db.WithContext(ctx).Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	if err != nil {
		// Consider using structured logging instead of fmt.Println
		return nil, err
	}
	result := make([]*domain.Product, len(list))
	for i, p := range list {
		result[i] = toDomain(p)
	}
	return result, nil
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	m := &model.ProductModel{
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProductRepository) Get(ctx context.Context, id int64) (*domain.Product, error) {
	var p model.ProductModel
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDataNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return toDomain(&p), nil
}

func toDomain(m *model.ProductModel) *domain.Product {
	return &domain.Product{
		ID:        m.ID,
		Name:      m.Name,
		Price:     m.Price,
		Stock:     m.Stock,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *ProductRepository) DeductStock(ctx context.Context, productId int64, count int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var m model.ProductModel

		// ① 行级悲观锁
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&m, productId).Error; err != nil {
			return err
		}

		// ② 库存校验（必须在事务内）
		if m.Stock < count {
			return errno.ProductErrNoEnoughStock
		}

		// ③ 原子更新
		if err := tx.Model(&m).
			Update("stock", m.Stock-count).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ProductRepository) DeductStockOptimistic(ctx context.Context, productId int64, count int64) (bool, error) {
	var product model.ProductModel
	err := r.db.WithContext(ctx).Where("id = ?", productId).First(&product).Error
	if err != nil {
		return false, err
	}
	if product.Stock < 1 {
		return false, errno.ProductErrNoEnoughStock
	}
	res := r.db.Model(&product).
		Where("id = ? AND version = ?", product.ID, product.Version).
		Updates(map[string]interface{}{
			"stock":   product.Stock - count,
			"version": product.Version + 1,
		})

	// ③ 原子更新 不使用额外字段
	/*res := r.db.WithContext(ctx).
	Model(&model.ProductModel{}).
	Where("id = ? AND stock >= ?", productId, count).
	Update("stock", gorm.Expr("stock - ?", count))*/

	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		// 竞争失败 or 库存不足
		return false, nil
	}

	return true, nil
}
