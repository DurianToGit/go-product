package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"product-service/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				response.Error(c, 50000, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
