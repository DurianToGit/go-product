package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"product-service/internal/domain"
	"product-service/internal/repository"
	"testing"
)

func TestProductService_Create(t *testing.T) {
	repo := repository.NewMockProductRepo()
	svc := NewProductService(repo)

	err := svc.CreateProduct(&domain.Product{ID: 1, Name: "Test"})
	assert.NoError(t, err)
}

func TestProductService_GetProduct(t *testing.T) {
	repo := repository.NewMockProductRepo()
	svc := NewProductService(repo)
	product, err := svc.GetProduct(1)
	fmt.Println(product)
	assert.Error(t, err)
}
