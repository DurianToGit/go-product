package userclient

import "context"

type UserProfile struct {
	ID       int64
	Username string
}
type UserInfo struct {
	ID        int64
	Username  string
	CreatedAt string
	UpdatedAt string
}

type ListQuery struct {
	Page     int64
	PageSize int64
	Keyword  string
	Status   int32
}

type Client interface {
	Register(ctx context.Context, username, password string) (*UserProfile, error)
	Login(ctx context.Context, username, password string) (string, error)
	GetUser(ctx context.Context, userID int64) (*UserProfile, error)
	UpdateUser(ctx context.Context, userID int64, username, password string) error
	UpdatePassword(ctx context.Context, username string, oldPassword, newPassword string) error
	ListUsers(ctx context.Context, query *ListQuery) ([]*UserInfo, error)
}
