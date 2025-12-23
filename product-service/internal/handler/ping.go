package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/response"
)

type CreateProductReq struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "pong",
	})
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")

	response.Success(c, gin.H{
		"id": id,
	})
}

func CreateProduct(c *gin.Context) {
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 40001, err.Error())
		return
	}
	response.Success(c, req)
}
