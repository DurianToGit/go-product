package domain

type Product struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Stock     int64  `json:"stock"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListReq struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}
