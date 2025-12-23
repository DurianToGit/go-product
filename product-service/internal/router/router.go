package router

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/handler"
)

func NewRouter(h *handler.ProductHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/products", h.List)
	r.POST("/products", h.Create)
	r.GET("/products/:id", h.Get)

	return r
}

func Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("ping", handler.Ping)

	api.GET("/products/:id", handler.GetProduct)
	api.POST("/products", handler.CreateProduct)
}
