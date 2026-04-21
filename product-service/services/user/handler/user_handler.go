package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"product-service/internal/errno"
	"product-service/pkg/response"
	"product-service/pkg/utils"
	"product-service/services/user/service"
	"strconv"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

func (h *UserHandler) List(c *gin.Context) {
	var req UserListReq
	if !utils.BindAndValidateByQuery(c, &req) {
		return
	}
	data, total, err := h.svc.List(c, req.ToDto())
	if err != nil {
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	result := response.ResultData{
		List: data,
		Mata: &response.Meta{
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	response.Success(c, result)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerReq
	if !utils.BindAndValidateByJSON(c, &req) {
		return
	}
	user, err := h.svc.Register(c, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errno.UsernameAlreadyExist) {
			response.ErrorWithErrno(c, errno.UsernameAlreadyExist)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req registerReq
	if !utils.BindAndValidateByJSON(c, &req) {
		return
	}
	token, err := h.svc.Login(c, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errno.UsernameNotFound) {
			response.ErrorWithErrno(c, errno.UsernameNotFound)
			return
		}
		if errors.Is(err, errno.UserErrPasswordIncorrect) {
			response.ErrorWithErrno(c, errno.UserErrPasswordIncorrect)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, gin.H{
		"token": token,
	})
}

func (h *UserHandler) Profile(c *gin.Context) {
	userId := utils.GetUserID(c)
	user, err := h.svc.GetByUserId(c, userId)
	if err != nil {
		if errors.Is(err, errno.UserErrNotFound) {
			response.ErrorWithErrno(c, errno.UserErrNotFound)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) Profile2(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	user, err := h.svc.GetByUserId(c, userId)
	if err != nil {
		if errors.Is(err, errno.UserErrNotFound) {
			response.ErrorWithErrno(c, errno.UserErrNotFound)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) UserInfo(c *gin.Context) {
	username := c.Query("username")
	user, err := h.svc.GetByUsername(c, username)
	if err != nil {
		if errors.Is(err, errno.UserErrNotFound) {
			response.ErrorWithErrno(c, errno.UserErrNotFound)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req updateUserReq
	if !utils.BindAndValidateByJSON(c, &req) {
		return
	}
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userId <= 0 {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}
	if err := h.svc.Update(c, userId, req.ToDto()); err != nil {
		if errors.Is(err, errno.UserErrNotFound) {
			response.ErrorWithErrno(c, errno.UserErrNotFound)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	var req updatePasswordReq
	if !utils.BindAndValidateByJSON(c, &req) {
		return
	}
	if err := h.svc.UpdatePassword(c, req.Username, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, errno.UserErrNotFound) {
			response.ErrorWithErrno(c, errno.UserErrNotFound)
			return
		}
		if errors.Is(err, errno.UserOldPasswordIncorrect) {
			response.ErrorWithErrno(c, errno.UserOldPasswordIncorrect)
			return
		}
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}
	response.Success(c, nil)
}
