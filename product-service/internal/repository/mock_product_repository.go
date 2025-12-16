package repository

import (
	"errors"
	"product-service/internal/domain"
)

type MockProductRepo struct {
	Products map[uint]*domain.Product
}

func NewMockProductRepo() *MockProductRepo {
	return &MockProductRepo{
		Products: make(map[uint]*domain.Product),
	}
}

func (m *MockProductRepo) Create(p *domain.Product) error {
	m.Products[p.ID] = p
	return nil
}

func (m *MockProductRepo) Get(id uint) (*domain.Product, error) {
	if p, ok := m.Products[id]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}
