package usergateway

import "github.com/gin-gonic/gin"

// InitPublicRouter 注册无需鉴权的公开路由
func InitPublicRouter(rg *gin.RouterGroup, h *Handler) {
	group := rg.Group("/users")
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
}

// InitRouter 注册需要鉴权的路由
func InitRouter(rg *gin.RouterGroup, h *Handler) {
	group := rg.Group("/users")
	group.GET("/:id", h.GetUser)
	group.PUT("/:id", h.UpdateUser)
	group.PUT("/password", h.UpdatePassword)
	group.GET("", h.ListUsers)
}
