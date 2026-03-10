package middleware

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/auth"
	"product-service/internal/errno"
	"product-service/pkg/response"
	"strings"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorWithErrno(c, errno.Unauthorized)
			c.Abort()
			return
		}
		// 解析 Authorization 头，格式应该是 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		// 检查是否为 Bearer 格式且包含 token
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorWithErrno(c, errno.Unauthorized)
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.ErrorWithErrno(c, errno.Unauthorized)
			c.Abort()
			return
		}
		// 把用户信息塞进 context
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
