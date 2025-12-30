package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/middleware"
)

func InitUserRouter(r *gin.RouterGroup, h *handler.UserHandler) {
	group := r.Group("user")
	{
		group.POST("register", h.Register)
		group.POST("login", h.Login)
		group.GET("userinfo", h.UserInfo)
	}
	authGroup := r.Group("user")
	authGroup.Use(middleware.Auth())
	{
		authGroup.GET("profile", h.Profile)
		authGroup.PUT("update_password", h.UpdatePassword)
	}
}
