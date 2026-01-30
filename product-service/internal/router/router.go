package router

import (
	"github.com/gin-gonic/gin"
	"product-service/api/handler"
	"product-service/internal/bootstrap"
	"product-service/internal/middleware"
	"product-service/pkg/ratelimit"
	"product-service/pkg/redis"
	"time"
)

func Register(r *gin.Engine, app *bootstrap.App) {
	// 链路追踪
	// r.Use(otelgin.Middleware("product-service"))
	// 分布式限流
	limiter := ratelimit.NewRedisLimiter(redis.Client)
	r.Use(
		middleware.DistributedRateLimit(
			limiter,
			middleware.DistributedRateLimitConfig{
				GlobalLimit:  200,
				GlobalWindow: time.Second,
				UserLimit:    5,
				UserWindow:   10 * time.Second,
			},
		),
	)
	api := r.Group("/api")
	api.GET("ping", handler.Ping)

	InitUserRouter(api, app.UserHandler)

	InitProductRouter(api, app.ProductHandler)

	InitOrderRouter(api, app.OrderHandler)
}
