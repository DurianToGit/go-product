package userclient

import (
	"context"

	userService "product-service/services/user/service"
)

type LocalClient struct {
	svc *userService.UserService
}

func (c *LocalClient) Register(ctx context.Context, username, password string) (*UserProfile, error) {
	// TODO implement me
	panic("implement me")
}

func (c *LocalClient) Login(ctx context.Context, username, password string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (c *LocalClient) GetUser(ctx context.Context, userID int64) (*UserProfile, error) {
	// TODO implement me
	panic("implement me")
}

func (c *LocalClient) UpdateUser(ctx context.Context, userID int64, nickname string, password string) error {
	// TODO implement me
	panic("implement me")
}

func (c *LocalClient) UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	// TODO implement me
	panic("implement me")
}

func (c *LocalClient) ListUsers(ctx context.Context, query *ListQuery) ([]*UserInfo, error) {
	// TODO implement me
	panic("implement me")
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
	}, nil
}
