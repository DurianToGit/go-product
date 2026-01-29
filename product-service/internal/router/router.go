package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"product-service/api/handler"
	"product-service/internal/bootstrap"
)

func Register(r *gin.Engine, app *bootstrap.App) {
	r.Use(otelgin.Middleware("product-service"))
	api := r.Group("/api")
	api.GET("ping", handler.Ping)

	InitUserRouter(api, app.UserHandler)

	InitProductRouter(api, app.ProductHandler)

	InitOrderRouter(api, app.OrderHandler)
}
