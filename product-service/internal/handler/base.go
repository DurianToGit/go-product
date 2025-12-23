package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/errno"
	"product-service/internal/response"
)

func BindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return false
	}
	return true
}
