package cache

import (
	"product-service/internal/domain"
	"sync"
)

type ProductCache struct {
	mu   sync.RWMutex
	data map[int64]domain.Product
}

func NewProductCache() *ProductCache {
	return &ProductCache{
		data: make(map[int64]domain.Product),
	}
}

func (p *ProductCache) Set(product domain.Product) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[product.ID] = product
}

func (p *ProductCache) Get(id int64) (domain.Product, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	product, ok := p.data[id]
	return product, ok
}

func (p *ProductCache) Delete(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.data, id)
}

func (p *ProductCache) List() []domain.Product {
	p.mu.RLock()
	defer p.mu.RUnlock()

	products := make([]domain.Product, 0, len(p.data))
	for _, product := range p.data {
		products = append(products, product)
	}
	return products
}
