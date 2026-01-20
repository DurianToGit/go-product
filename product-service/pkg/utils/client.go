package utils

import (
	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}
