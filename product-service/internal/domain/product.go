package domain

import "errors"

type Product struct {
	ID    int64  `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}

type ListReq struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"page_size"`
}

var (
	ErrProductNotFound = errors.New("product not found")
)
