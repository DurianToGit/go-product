package userclient

import "context"

type UserProfile struct {
	ID       int64
	Username string
	Nickname string
}

type Client interface {
	GetProfile(ctx context.Context, userID int64) (*UserProfile, error)
}
