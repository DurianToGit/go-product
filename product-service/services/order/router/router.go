package router

import (
	"github.com/gin-gonic/gin"
	"product-service/services/order/handler"
)

func InitOrderRouter(r *gin.RouterGroup, h *handler.OrderHandler) {
	group := r.Group("orders")
	{
		group.POST("", h.Create)
		group.POST("pay", h.Pay)
	}
}
