package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/response"
)

type CreateProductReq struct {
	Name  string `json:"name" binding:"required"`
	Price int    `json:"price" binding:"required,gt=0"`
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
	if !BindAndValidate(c, &req) {
		return
	}
	response.Success(c, req)
}
