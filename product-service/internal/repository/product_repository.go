package repository

import (
	"product-service/internal/domain"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db}
}

func (r *ProductRepository) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *ProductRepository) Get(id uint) (*domain.Product, error) {
	var p domain.Product
	err := r.db.First(&p, id).Error
	return &p, err
}
