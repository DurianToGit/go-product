package model

import (
	"gorm.io/gorm"
)

type ProductModel struct {
	BaseModel
	Name      string    `gorm:"type:varchar(128);not null"`
	Price     int64     `gorm:"not null;default:0"`
	Stock     int64     `gorm:"not null;default:0"`
	Version   int       `gorm:"not null;default:1"`
	CreatorID int64     `gorm:"not null;index:idx_creator_id"`
	Creator   UserModel `gorm:"foreignKey:CreatorID;references:ID"`
}

func (ProductModel) TableName() string {
	return "products"
}

func (p *ProductModel) BeforeCreate(tx *gorm.DB) error {
	if p.Version <= 0 {
		p.Version = 1
	}
	return nil
}

func (p *ProductModel) BeforeSave(tx *gorm.DB) error {
	// p.Name = strings.TrimSpace(p.Name)
	// if p.Name == "" {
	// 	return errno.ProductErrNameEmpty
	// }
	return nil
}
