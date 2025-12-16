package service

import (
	"product-service/internal/domain"
	"product-service/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetProducts(page, pageSize int) ([]*domain.Product, error) {
	return s.repo.List(page, pageSize)
}

func (s *ProductService) CreateProduct(p *domain.Product) error {
	return s.repo.Create(p)
}

func (s *ProductService) GetProduct(id uint) (*domain.Product, error) {
	return s.repo.Get(id)
}
