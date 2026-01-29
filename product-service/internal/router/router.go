package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"product-service/api/handler"
	"product-service/internal/bootstrap"
)

func Register(r *gin.Engine, app *bootstrap.App) {
	api := r.Group("/api")
	r.Use(otelgin.Middleware("product-service"))
	api.GET("ping", handler.Ping)

	InitUserRouter(api, app.UserHandler)

	InitProductRouter(api, app.ProductHandler)

	InitOrderRouter(api, app.OrderHandler)
}
