package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"product-service/internal/auth"
	"product-service/internal/errno"
	"product-service/internal/response"
	"product-service/internal/service"
)

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type updatePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterReq
	if !BindAndValidateByJSON(c, &req) {
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
	if !BindAndValidateByJSON(c, &req) {
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
	user, err := service.GetUserInfo(userId.(int64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorWithErrno(c, errno.ErrUserNotFound)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, user)
}

func UpdatePassword(c *gin.Context) {
	var req updatePasswordReq
	if !BindAndValidateByJSON(c, &req) {
		return
	}
	userId, _ := c.Get("user_id")
	if err := service.UpdatePassword(userId.(int64), req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, errno.OldPasswordIncorrect) {
			response.ErrorWithErrno(c, errno.OldPasswordIncorrect)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, gin.H{})
}
