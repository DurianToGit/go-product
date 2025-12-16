package api

import (
	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	repo := repository.NewProductRepository(db)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	r.POST("/products", h.Create)
	r.GET("/products/:id", h.Get)

	return r
}
