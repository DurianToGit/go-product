package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/auth"
	"product-service/internal/errno"
	"product-service/internal/response"
)

func Login(c *gin.Context) {
	// 模拟用户登录成功
	userId := int64(10008)
	token, err := auth.GenerateToken(userId)
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, gin.H{
		"token": token,
	})
}

func Profile(c *gin.Context) {
	userId, _ := c.Get("user_id")
	response.Success(c, gin.H{
		"user_id": userId,
	})
}
