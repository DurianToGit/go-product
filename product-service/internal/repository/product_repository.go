package repository

import (
	"errors"
	"fmt"
	"product-service/internal/domain"

	"gorm.io/gorm"
)

type ProductRepository interface {
	List(page, pageSize int) ([]*domain.Product, error)
	Create(p *domain.Product) error
	Get(id int64) (*domain.Product, error)
}

type MysqlProductRepo struct {
	db *gorm.DB
}

func NewMysqlProductRepo(db *gorm.DB) *MysqlProductRepo {
	return &MysqlProductRepo{db: db}
}

func (r *MysqlProductRepo) List(page, pageSize int) ([]*domain.Product, error) {
	fmt.Println("参数page=", page, "pageSize=", pageSize)
	// Add input validation
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // default page size
	}

	var list []*domain.Product
	err := r.db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	if err != nil {
		// Consider using structured logging instead of fmt.Println
		return nil, err
	}
	return list, nil
}

func (r *MysqlProductRepo) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *MysqlProductRepo) Get(id int64) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return &p, nil
}

// func (r *ProductRepository) Create(p *domain.Product) error {
// 	return r.db.Create(p).Error
// }
//
// func (r *ProductRepository) Get(id int64) (*domain.Product, error) {
// 	var p domain.Product
// 	err := r.db.First(&p, id).Error
// 	return &p, err
// }
