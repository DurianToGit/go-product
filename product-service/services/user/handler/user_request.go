package handler

import "product-service/services/user/dto"

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type updatePasswordReq struct {
	Username    string `json:"Username" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserListReq struct {
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type updateUserReq struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

func (r *UserListReq) ToDto() *dto.UserQuery {
	return &dto.UserQuery{
		Keyword:  r.Keyword,
		Status:   r.Status,
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

func (r *updateUserReq) ToDto() *dto.UserUpdate {
	return &dto.UserUpdate{
		Username: r.Username,
		Password: r.Password,
	}
}
