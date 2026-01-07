package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/middleware"
	"product-service/pkg/utils"
)

func InitUserRouter(r *gin.RouterGroup, h *handler.UserHandler) {
	group := r.Group("users")
	{
		group.POST("register", h.Register)
		group.POST("login", middleware.RateLimit("users:rl:ip", 5.0, 10, utils.GetClientIP), h.Login)
		group.GET("userinfo", h.UserInfo)
		group.GET("", h.List)
		group.GET(":id", h.Profile2)
		group.PATCH("/:id", h.Update)
	}
	authGroup := r.Group("users")
	authGroup.Use(middleware.Auth())
	{
		authGroup.GET("profile", h.Profile)
		authGroup.PUT("update_password", h.UpdatePassword)
	}
}
