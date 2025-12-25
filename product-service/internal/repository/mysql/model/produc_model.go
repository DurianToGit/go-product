package model

import (
	"product-service/pkg/db"
	"time"
)

type ProductModel struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:128;not null"`
	Price     int64  `gorm:"not null"`
	Stock     int64  `gorm:"not null"`
	Version   int64  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProductModel) TableName() string {
	return "products"
}

func (p *ProductModel) Create() error {
	return db.DB.Create(p).Error
}
func (p *ProductModel) Update() error {
	return db.DB.Save(p).Error
}
func (p *ProductModel) Delete() error {
	return db.DB.Delete(p).Error
}
