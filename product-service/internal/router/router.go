package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/bootstrap"
)

func Register(r *gin.Engine, app *bootstrap.App) {
	api := r.Group("/api")
	api.GET("ping", handler.Ping)

	InitUserRouter(api, app.UserHandler)

	InitProductRouter(api, app.ProductHandler)

	InitOrderRouter(api, app.OrderHandler)
}
