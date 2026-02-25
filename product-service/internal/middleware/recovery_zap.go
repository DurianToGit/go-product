package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"product-service/internal/logger"
	"product-service/internal/response"
)

func RecoveryZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error("panic_recovered",
					zap.Any("panic", r),
					zap.String("request_id", GetRequestID(c)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Stack("stack"),
				)
				response.Error(c, 50000, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
