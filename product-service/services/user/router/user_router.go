package router

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/middleware"
	"product-service/pkg/utils"
	"product-service/services/user/handler"
)

func InitAuthRouter(r *gin.RouterGroup, h *handler.UserHandler) {
	r.POST("register", h.Register)
	r.POST("login", middleware.RateLimit("users:rl:ip", 5.0, 10, utils.GetClientIP), h.Login)
}

func InitUserRouter(r *gin.RouterGroup, h *handler.UserHandler) {
	group := r.Group("users")
	{
		group.GET("profile", h.Profile)
		group.PUT("update_password", h.UpdatePassword)
		group.GET("userinfo", h.UserInfo)
		group.GET("", h.List)
		group.GET(":id", h.Profile2)
		group.PATCH("/:id", h.Update)
	}
}
