package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
)

func InitProductRouter(r *gin.RouterGroup, h *handler.ProductHandler) {
	group := r.Group("products")
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.GET("/:id/detail", h.GetWithCreator)
		group.GET("/deduct/:id", h.DuctStock)
		group.GET("/deduct_op/:id", h.DuctStockOptimistic)

	}
}
