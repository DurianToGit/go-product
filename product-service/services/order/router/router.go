package router

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/middleware"
	"product-service/services/order/handler"
)

func InitOrderRouter(r *gin.RouterGroup, h *handler.OrderHandler) {
	group := r.Group("orders").Use(middleware.Auth())
	{
		group.POST("", h.Create)
	}
}
