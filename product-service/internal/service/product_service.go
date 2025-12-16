package service

import (
	"product-service/internal/domain"
	"product-service/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo}
}

func (s *ProductService) CreateProduct(p *domain.Product) error {
	return s.repo.Create(p)
}

func (s *ProductService) GetProduct(id uint) (*domain.Product, error) {
	return s.repo.Get(id)
}
