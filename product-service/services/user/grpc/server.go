package grpc

import (
	"context"
	"fmt"
	"product-service/pkg/pb/userpb"
	"product-service/services/user/dto"
	"product-service/services/user/service"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
	svc *service.UserService
}

func NewServer(svc *service.UserService) *Server {
	return &Server{
		svc: svc,
	}
}

func (s *Server) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	resp, err := s.svc.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &userpb.RegisterResponse{
		Id:       resp.ID,
		Username: resp.Username,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	token, err := s.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &userpb.LoginResponse{
		Token: token,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	if req.GetId() != 0 {
		resp, err := s.svc.GetByUserId(ctx, req.GetId())
		if err != nil {
			return nil, err
		}
		return &userpb.GetUserResponse{
			Id:        resp.ID,
			Username:  resp.Username,
			CreatedAt: resp.CreatedAt,
			UpdatedAt: resp.UpdatedAt,
		}, nil
	}
	if req.GetUsername() != "" {
		resp, err := s.svc.GetByUsername(ctx, req.GetUsername())
		if err != nil {
			return nil, err
		}
		return &userpb.GetUserResponse{
			Id:        resp.ID,
			Username:  resp.Username,
			CreatedAt: resp.CreatedAt,
			UpdatedAt: resp.UpdatedAt,
		}, nil
	}
	return nil, fmt.Errorf("no id or username provided")
}
func (s *Server) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	err := s.svc.Update(ctx, req.GetId(), &dto.UserUpdate{
		Username: &req.Username,
		Password: &req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserResponse{
		Success: true,
	}, nil
}
func (s *Server) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	status := int(req.GetStatus())
	list, total, err := s.svc.List(ctx, &dto.UserQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Keyword:  req.Keyword,
		Status:   &status,
	})
	if err != nil {
		return nil, err
	}
	users := make([]*userpb.GetUserResponse, 0, len(list))
	for _, user := range list {
		users = append(users, &userpb.GetUserResponse{
			Id:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return &userpb.ListUsersResponse{
		List:  users,
		Total: total,
	}, nil
}
func (s *Server) UpdatePassword(ctx context.Context, req *userpb.UpdatePasswordRequest) (*userpb.UpdatePasswordResponse, error) {
	err := s.svc.UpdatePassword(ctx, req.GetUsername(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		return nil, err
	}
	return &userpb.UpdatePasswordResponse{
		Success: true,
	}, nil
}
