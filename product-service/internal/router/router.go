package router

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/handler"
	"product-service/internal/middleware"
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
	// 登录
	api.POST("/login", handler.Login)
	authApi := api.Group("/")
	authApi.Use(middleware.Auth())
	{
		authApi.GET("profile", handler.Profile)
	}

	api.GET("/products/:id", handler.GetProduct)
	api.POST("/products", handler.CreateProduct)

}
