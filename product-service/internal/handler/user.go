package handler

import (
	"github.com/gin-gonic/gin"
	"product-service/internal/auth"
	"product-service/internal/errno"
	"product-service/internal/response"
	"product-service/internal/service"
)

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterReq
	if !BindAndValidate(c, &req) {
		return
	}
	if err := service.Register(req.Username, req.Password); err != nil {
		response.Error(c, 50000, err.Error())
		return
	}

	response.Success(c, gin.H{})
}

func Login(c *gin.Context) {
	var req RegisterReq
	if !BindAndValidate(c, &req) {
		return
	}
	user, err := service.Login(req.Username, req.Password)
	if err != nil {
		response.ErrorWithErrno(c, errno.Unauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID)
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
