package domain

type Product struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}

type ListReq struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"page_size"`
}
