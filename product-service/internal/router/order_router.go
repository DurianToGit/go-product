package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/middleware"
)

func InitOrderRouter(r *gin.RouterGroup, h *handler.OrderHandler) {
	group := r.Group("orders").Use(middleware.Auth())
	{
		group.POST("", h.Create)
	}
}
