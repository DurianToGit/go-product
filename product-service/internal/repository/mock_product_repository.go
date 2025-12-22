package repository

import (
	"errors"
	"product-service/internal/domain"
)

type MockProductRepo struct {
	Products map[int64]*domain.Product
}

func NewMockProductRepo() *MockProductRepo {
	return &MockProductRepo{
		Products: make(map[int64]*domain.Product),
	}
}

func (m *MockProductRepo) List(page, pageSize int) ([]*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockProductRepo) Create(p *domain.Product) error {
	m.Products[p.ID] = p
	return nil
}

func (m *MockProductRepo) Get(id int64) (*domain.Product, error) {
	if p, ok := m.Products[id]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}
