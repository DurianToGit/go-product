package middleware

import (
	"product-service/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		// 如果 handler 里用 c.Error(err)，会挂在 c.Errors
		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		fields := []zap.Field{
			zap.String("request_id", GetRequestID(c)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Int("bytes", size),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("ua", c.Request.UserAgent()),
		}
		if errMsg != "" {
			fields = append(fields, zap.String("error", errMsg))
			logger.L().Warn("http_request", fields...)
			return
		}

		// 4xx/5xx 也可以提升等级
		if status >= 500 {
			logger.L().Error("http_request", fields...)
		} else if status >= 400 {
			logger.L().Warn("http_request", fields...)
		} else {
			logger.L().Info("http_request", fields...)
		}
	}
}
