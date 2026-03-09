package domain

type Product struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Stock     int64  `json:"stock"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

const ProductEventTypeStockDeducted = "stock_deducted"
const ProductEventTypeRestockDeducted = "restock_deducted"

// 数据库不存在
const DataCacheNil = "__nil__"
