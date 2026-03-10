package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/pkg/response"
)

func Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "pong",
	})
}
