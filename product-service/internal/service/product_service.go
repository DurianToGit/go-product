package service

import (
	"context"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/errno"
	"product-service/internal/repository"
	"time"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetProducts(ctx context.Context, q *dto.ProductQuery) ([]*domain.Product, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *ProductService) CreateProduct(ctx context.Context, p *domain.Product) error {
	return s.repo.Create(ctx, p)
}

func (s *ProductService) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProductService) GetProductWithCreator(ctx context.Context, id int64) (*domain.Product, *domain.User, error) {
	return s.repo.GetWithCreator(ctx, id)
}

func (s *ProductService) DeductStock(ctx context.Context, productId int64, count int64) error {
	if count <= 0 {
		return errno.ProductErrInvalidStock
	}
	return s.repo.DeductStock(ctx, productId, count)
}

func (s *ProductService) DeductStockOptimistic(ctx context.Context, productId int64, count int64) error {
	if count <= 0 {
		return errno.ProductErrInvalidStock
	}
	// 加入重试机制
	const maxRetry = 3
	for i := 0; i < maxRetry; i++ {
		ok, err := s.repo.DeductStockOptimistic(ctx, productId, count)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return errno.ProductErrStockNotEnough
}

func (s *ProductService) ValidateAsync(p *domain.Product) error {
	ch := make(chan error)

	go func() {
		// 模拟耗时校验
		time.Sleep(100 * time.Millisecond)
		ch <- nil
	}()

	return <-ch
}
