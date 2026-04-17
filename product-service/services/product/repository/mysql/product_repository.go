package mysql

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"product-service/internal/errno"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/services/product/domain"
	"product-service/services/product/dto"
	"product-service/services/product/repository/mysql/model"
	"time"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, q *dto.ProductQuery) ([]*domain.Product, int64, error) {
	var (
		list  []*model.ProductModel
		total int64
	)
	page := q.Page
	pageSize := q.PageSize

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&model.ProductModel{})
	db = db.Preload("Creator")
	if q.Keyword != "" {
		db = db.Where("products.name LIKE ?", "%"+q.Keyword+"%")
	}
	if q.CreatorUsername != "" {
		db = db.Joins("Creator").Where("Creator.username LIKE ?", "%"+q.CreatorUsername+"%")
	}
	if q.Status != nil {
		db = db.Where("products.status = ?", *q.Status)
	}
	if q.MinPrice != nil {
		db = db.Where("products.price >= ?", *q.MinPrice)
	}
	if q.MaxPrice != nil {
		db = db.Where("products.price <= ?", *q.MaxPrice)
	}
	db.Count(&total)
	err := db.Order("products.id DESC").Limit(pageSize).Offset(offset).Find(&list).Error
	if err != nil {
		// Consider using structured logging instead of fmt.Println
		return nil, 0, err
	}
	result := make([]*domain.Product, len(list))
	for i, p := range list {
		result[i] = toDomain(p)
	}
	return result, total, nil
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

func (r *ProductRepository) GetTx(ctx context.Context, tx *gorm.DB, id int64) (*domain.Product, error) {
	var p model.ProductModel
	if err := tx.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDataNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return toDomain(&p), nil
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

func (r *ProductRepository) DeductStockAtomic(ctx context.Context, productID, count int64) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&model.ProductModel{}).
		Where("id = ? AND stock >= ?", productID, count).
		Update("stock", gorm.Expr("stock - ?", count))

	if res.Error != nil {
		return false, res.Error
	}
	if err := redis.Client.DecrBy(ctx, rediskeys.ProductStockKey(productID), count).Err(); err != nil {
		logger.L().Error("扣减 redis 库存失败",
			zap.Int64("product_id", productID),
			zap.Int64("count", count),
			zap.Error(err),
		)
	}

	return res.RowsAffected == 1, nil
}

func (r *ProductRepository) RestoreStock(ctx context.Context, productID int64, count int64) error {
	tx := r.db.WithContext(ctx).
		Model(&model.ProductModel{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", count))

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errno.ErrDataNotFound
	}

	if err := redis.Client.IncrBy(ctx, rediskeys.ProductStockKey(productID), count).Err(); err != nil {
		logger.L().Error("恢复 redis 库存失败",
			zap.Int64("product_id", productID),
			zap.Int64("count", count),
			zap.Error(err),
		)
	}

	return nil
}

func (r *ProductRepository) ConsumeStockDeductEvent(
	ctx context.Context,
	eventSource, eventID string,
	productID, count int64,
	eventType string,
) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 幂等记录（先插入）
		rec := &model.ProductEventConsumedModel{
			EventSource: eventSource,
			EventId:     eventID,
			ProductId:   productID,
			Count:       count,
			EventType:   eventType,
			CreatedAt:   time.Now().Unix(),
		}

		if err := tx.Create(rec).Error; err != nil {
			// duplicate -> 已处理过，直接返回 nil，让上层 ACK
			if isDuplicateKeyError(err) {
				return nil
			}
			return err
		}

		// 2) 扣 MySQL 库存（原子）
		res := tx.Model(&model.ProductModel{}).
			Where("id = ? AND stock >= ?", productID, count).
			Update("stock", gorm.Expr("stock - ?", count))

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Redis 扣了但 DB 不够：一致性异常
			return errno.ProductErrNoEnoughStock
		}

		return nil
	})
}

func (r *ProductRepository) ConsumeRestockDeductEvent(
	ctx context.Context,
	eventSource, eventID string,
	productID, count int64,
	eventType string,
) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 幂等记录（先插入）
		rec := &model.ProductEventConsumedModel{
			EventSource: eventSource,
			EventId:     eventID,
			ProductId:   productID,
			Count:       count,
			EventType:   eventType,
			CreatedAt:   time.Now().Unix(),
		}

		if err := tx.Create(rec).Error; err != nil {
			// duplicate -> 已处理过，直接返回 nil，让上层 ACK
			if isDuplicateKeyError(err) {
				return nil
			}
			return err
		}

		// 2) 扣 MySQL 库存（原子）
		res := tx.Model(&model.ProductModel{}).
			Where("id = ?", productID).
			Update("stock", gorm.Expr("stock + ?", count))

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Redis 加了 但是产品不存在？
			return errno.ErrDataNotFound
		}
		err := redis.Client.IncrBy(ctx, rediskeys.ProductStockKey(productID), int64(count)).Err()
		if err != nil {
			// 此处错误忽略 现在这个项目内并没有做定期的库存重写到redis中，所以这里忽略
			// 在常规开发中，这里直接删除redis缓存，而不是修改，即使是修改也是可以允许暂时的不一致，定期重写商品库存缓存即可
			// 或者此类 统一用事务提交到outbox表，统一处理，这里就不在写redis了
			logger.L().Info("redis 恢复库存失败", zap.Error(err), zap.Int64("product_id", productID), zap.Int64("count", count))
			// return err
		}

		return nil
	})
}

func isDuplicateKeyError(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
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
