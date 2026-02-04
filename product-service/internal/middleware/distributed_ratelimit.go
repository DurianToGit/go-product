package middleware

import (
	"fmt"
	"log"
	"product-service/internal/errno"
	"product-service/internal/response"
	"product-service/pkg/breaker"
	"time"

	"github.com/gin-gonic/gin"
	"product-service/pkg/ratelimit"
)

type DistributedRateLimitConfig struct {
	GlobalLimit  int64
	GlobalWindow time.Duration

	UserLimit  int64
	UserWindow time.Duration
}

func DistributedRateLimit(limiter *ratelimit.RedisLimiter, cfg DistributedRateLimitConfig) gin.HandlerFunc {

	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		// ---------- 全局限流 ----------
		gKey := fmt.Sprintf(
			"rl:global:%s:%d",
			route,
			time.Now().UnixNano()/int64(cfg.GlobalWindow),
		)

		ok, err := limiter.Allow(
			c.Request.Context(),
			gKey,
			cfg.GlobalLimit,
			cfg.GlobalWindow,
		)
		if err != nil {
			if breaker.IsCircuitOpen(err) {
				// 直接降级/快速返回
				response.ErrorWithErrno(c, errno.ErrDependencyUnavailable)
				log.Printf("依赖不可用: %v", err)
				c.Abort()
				return
			}
			response.ErrorWithErrno(c, errno.ErrDependencyUnavailable)
			c.Abort()
			return
		}
		if !ok {
			/*c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{"message": "global rate limit exceeded"},
			)*/
			response.ErrorWithErrno(c, errno.ErrTooManyRequests)
			log.Printf("全局限流已超: %v", err)
			c.Abort()
			return
		}

		// ---------- 用户限流 ----------
		if uid, exists := c.Get("user_id"); exists {
			userID := uid.(int64)

			uKey := fmt.Sprintf(
				"rl:user:%d:%s:%d",
				userID,
				route,
				time.Now().UnixNano()/int64(cfg.UserWindow),
			)

			ok, err := limiter.Allow(
				c.Request.Context(),
				uKey,
				cfg.UserLimit,
				cfg.UserWindow,
			)
			if err != nil {
				if breaker.IsCircuitOpen(err) {
					// 直接降级/快速返回
					response.ErrorWithErrno(c, errno.ErrDependencyUnavailable)
					log.Printf("依赖不可用1: %v", err)
					c.Abort()
					return
				}
				response.ErrorWithErrno(c, errno.ErrDependencyUnavailable)
				c.Abort()
				return
			}
			if !ok {
				// c.AbortWithStatusJSON(
				// 	http.StatusTooManyRequests,
				// 	gin.H{"message": "user rate limit exceeded"},
				// )
				response.ErrorWithErrno(c, errno.ErrUserRateLimitExceeded)
				log.Printf("用户限流已超: %v", err)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
