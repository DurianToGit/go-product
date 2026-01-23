package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/middleware"
	"product-service/pkg/utils"
)

func InitProductRouter(r *gin.RouterGroup, h *handler.ProductHandler) {
	group := r.Group("products")
	{
		group.GET("", middleware.RateLimit("products:rl:ip", 20.0, 40, utils.GetClientIP), h.List)
		group.GET("/search", middleware.RateLimit("products:search:rl:ip", 20.0, 40, utils.GetClientIP), h.Search)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.GET("/:id/detail", h.GetWithCreator)
		group.POST("/deduct_seckill/:id", h.DuctStockSeckill)
		group.POST("/:id/stock/prewarm", h.PrewarmProductStock)
	}
}
