package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func Cost() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		cost := time.Since(start)
		// cost转为毫秒
		cost = cost / time.Millisecond
		log.Printf("[COST]%s %s %dms", c.Request.Method, c.Request.URL.Path, cost)
	}
}
