package dto

type ProductQuery struct {
	Keyword         string
	CreatorUsername string
	MinPrice        *int64
	MaxPrice        *int64
	Status          *int
	Page            int
	PageSize        int
}
