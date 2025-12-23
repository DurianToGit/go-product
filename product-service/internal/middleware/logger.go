package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[REQ]%s %s from %s", c.Request.Method, c.Request.URL.Path, c.Request.RemoteAddr)
		c.Next()
	}
}
