package usergateway

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"product-service/internal/client/userclient"
	"product-service/internal/errno"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/response"
)

type Handler struct {
	client userclient.Client
}

func NewHandler(client userclient.Client) *Handler {
	return &Handler{client: client}
}

// ---------- 公开接口（无需鉴权） ----------

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway register", zap.String("username", req.Username), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	user, err := h.client.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		logger.L().Error("gateway register failed", zap.Error(err), zap.String("username", req.Username))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway login", zap.String("username", req.Username), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	token, err := h.client.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		logger.L().Error("gateway login failed", zap.Error(err), zap.String("username", req.Username))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}

// ---------- 鉴权接口 ----------

func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway get user", zap.Int64("user_id", userID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	user, err := h.client.GetUser(c.Request.Context(), userID)
	if err != nil {
		logger.L().Error("gateway get user failed", zap.Error(err), zap.Int64("user_id", userID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

type UpdateUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway update user", zap.Int64("user_id", userID), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	if err := h.client.UpdateUser(c.Request.Context(), userID, req.Username, req.Password); err != nil {
		logger.L().Error("gateway update user failed", zap.Error(err), zap.Int64("user_id", userID))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, nil)
}

type UpdatePasswordReq struct {
	Username    string `json:"username" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *Handler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway update password", zap.String("username", req.Username), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	if err := h.client.UpdatePassword(c.Request.Context(), req.Username, req.OldPassword, req.NewPassword); err != nil {
		logger.L().Error("gateway update password failed", zap.Error(err), zap.String("username", req.Username))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	response.Success(c, nil)
}

type ListUsersReq struct {
	Page     int64  `form:"page" binding:"required,min=1"`
	PageSize int64  `form:"page_size" binding:"required,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   int32  `form:"status"`
}

func (h *Handler) ListUsers(c *gin.Context) {
	var req ListUsersReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorWithErrno(c, errno.InvalidParams)
		return
	}

	logger.L().Info("gateway list users", zap.Int64("page", req.Page), zap.Int64("page_size", req.PageSize), zap.String("request_id", grpcx.GetRequestIDFromContext(c.Request.Context())))

	users, err := h.client.ListUsers(c.Request.Context(), &userclient.ListQuery{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Status:   req.Status,
	})
	if err != nil {
		logger.L().Error("gateway list users failed", zap.Error(err))
		response.ErrorWithErrno(c, errno.ServerError)
		return
	}

	list := make([]gin.H, 0, len(users))
	for _, u := range users {
		list = append(list, gin.H{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		})
	}

	response.Success(c, gin.H{
		"list": list,
		"meta": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}
