package ordergateway

import "github.com/gin-gonic/gin"

func InitRouter(rg *gin.RouterGroup, h *Handler) {
	group := rg.Group("/orders")
	group.POST("", h.CreateOrder)
	group.GET("/:id", h.GetOrder)
	group.POST("/:id/cancel", h.CancelOrder)
	group.POST("/:id/pay", h.PayOrder)
}
