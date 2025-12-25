package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/middleware"
)

func NewRouter(h *handler.ProductHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/products", h.List)
	r.POST("/products", h.Create)
	r.GET("/products/:id", h.Get)

	return r
}

func Register(r *gin.Engine) *gin.RouterGroup {
	api := r.Group("/api")
	api.GET("ping", handler.Ping)

	userApi := api.Group("/user")
	// 登录
	userApi.POST("/register", handler.Register)
	// 注册
	userApi.POST("/login", handler.Login)

	authApi := userApi.Group("/")
	authApi.Use(middleware.Auth())
	{
		authApi.GET("profile", handler.Profile)
		authApi.PUT("update_password", handler.UpdatePassword)
	}

	// productApi := api.Group("/products")
	// productApi.GET("/:id", handler.GetProduct)
	// productApi.POST("/", handler.CreateProduct)
	// productApi.GET("/deduct", handler.DeductStock)
	return api
}
