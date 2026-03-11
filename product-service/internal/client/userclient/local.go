package userclient

import (
	"context"

	userService "product-service/services/user/service"
)

type LocalClient struct {
	svc *userService.UserService
}

func NewLocalClient(svc *userService.UserService) Client {
	return &LocalClient{svc: svc}
}

func (c *LocalClient) GetProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	u, err := c.svc.GetByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Username, // 当前结构暂无昵称字段
	}, nil
}
