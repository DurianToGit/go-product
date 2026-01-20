package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/errno"
	"product-service/internal/response"
)

func BindAndValidateByJSON(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return false
	}
	return true
}

func BindAndValidateByQuery(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return false
	}
	return true
}

func GetUserID(c *gin.Context) int64 {
	return c.GetInt64("user_id")
}
