package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}

func GetUserId(c *gin.Context) string {
	userId, ok := c.Get("user_id")
	if !ok {
		log.Println("GetUserId: user_id not found")
		return "caller"
	}
	return fmt.Sprint(userId)
}
