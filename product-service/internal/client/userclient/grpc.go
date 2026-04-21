package userclient

import (
	"context"
	"fmt"
	"product-service/pkg/pb/userpb"

	"google.golang.org/grpc"
)

type GRPCClient struct {
	client userpb.UserServiceClient
}

func (g *GRPCClient) Register(ctx context.Context, username, password string) (*UserProfile, error) {
	resp, err := g.client.Register(ctx, &userpb.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		ID:       resp.Id,
		Username: resp.Username,
	}, nil
}

func (g *GRPCClient) Login(ctx context.Context, username, password string) (string, error) {
	token, err := g.client.Login(ctx, &userpb.LoginRequest{
		Username: username,
		Password: password,
	})
	return token.Token, err
}

func (g *GRPCClient) GetUser(ctx context.Context, userID int64) (*UserProfile, error) {
	resp, err := g.client.GetUser(ctx, &userpb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		ID:       resp.Id,
		Username: resp.Username,
	}, nil
}

func (g *GRPCClient) UpdateUser(ctx context.Context, userID int64, username, password string) error {
	resp, err := g.client.UpdateUser(ctx, &userpb.UpdateUserRequest{
		Id:       userID,
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("update user failed")
	}
	return nil
}

func (g *GRPCClient) UpdatePassword(ctx context.Context, username string, oldPassword, newPassword string) error {
	resp, err := g.client.UpdatePassword(ctx, &userpb.UpdatePasswordRequest{
		Username:    username,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("update password failed")
	}
	return nil
}

func (g *GRPCClient) ListUsers(ctx context.Context, query *ListQuery) ([]*UserInfo, error) {
	resp, err := g.client.ListUsers(ctx, &userpb.ListUsersRequest{
		Page:     query.Page,
		PageSize: query.PageSize,
		Keyword:  query.Keyword,
		Status:   query.Status,
	})
	if err != nil {
		return nil, err
	}
	var list []*UserInfo
	for _, user := range resp.List {
		list = append(list, &UserInfo{
			ID:        user.Id,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return list, nil
}

func (g *GRPCClient) GetProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	resp, err := g.client.GetUser(ctx, &userpb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		ID:       resp.Id,
		Username: resp.Username,
	}, nil
}

func NewGRPCClient(conn *grpc.ClientConn) Client {
	return &GRPCClient{
		client: userpb.NewUserServiceClient(conn),
	}
}
