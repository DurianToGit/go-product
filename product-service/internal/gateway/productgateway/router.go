package productgateway

import "github.com/gin-gonic/gin"

func InitRouter(rg *gin.RouterGroup, h *Handler) {
	group := rg.Group("/products")
	group.GET("/:id", h.GetProduct)
	group.GET("/:id/stock", h.GetStock)
}
