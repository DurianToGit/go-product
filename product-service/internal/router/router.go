package router

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/bootstrap"
	"product-service/internal/gateway/productgateway"
	"product-service/internal/middleware"
	"product-service/pkg/health"
	"product-service/pkg/ratelimit"
	"product-service/pkg/redis"
	orderRouter "product-service/services/order/router"
	productRouter "product-service/services/product/router"
	userRouter "product-service/services/user/router"
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
	r.GET("health", health.Handler(&health.Checker{
		DB:    app.MysqlClient,
		Redis: app.RedisClient,
	}))
	api := r.Group("/api/v1")

	// 公共路由
	authGroup := api.Group("/auth")
	{
		userRouter.InitAuthRouter(authGroup, app.UserHandler)
	}
	if app.ProductGatewayHandler != nil {
		gateway := api.Group("/gateway")
		gateway.Use(middleware.Auth())
		{
			productgateway.InitRouter(gateway, app.ProductGatewayHandler)
		}
	}

	// 需要登录的业务路由
	biz := api.Group("")
	biz.Use(middleware.Auth())
	{
		productRouter.InitProductRouter(biz, app.ProductHandler)
		orderRouter.InitOrderRouter(biz, app.OrderHandler)
		userRouter.InitUserRouter(biz, app.UserHandler)
	}
}
