package grpc

import (
	"context"
	"product-service/pkg/pb/orderpb"
	orderservice "product-service/services/order/service"
)

type Server struct {
	orderpb.UnimplementedOrderServiceServer
	svc *orderservice.OrderService
}

func NewServer(svc *orderservice.OrderService) *Server {
	return &Server{
		svc: svc,
	}
}
func (s *Server) Create(ctx context.Context, req *orderpb.CreateRequest) (*orderpb.CreateResponse, error) {
	order, err := s.svc.Create(ctx, req.UserId, req.ProductId, int(req.Count), req.IdemKey)
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateResponse{
		Id:        order.ID,
		OrderNo:   order.OrderNo,
		UserId:    order.UserID,
		ProductId: order.ProductID,
		Count:     int64(order.Count),
		Amount:    order.Amount,
		Status:    order.Status,
		IdemKey:   order.IdemKey,
	}, nil
}

func (s *Server) Cancel(ctx context.Context, req *orderpb.CancelRequest) (*orderpb.CancelResponse, error) {
	err := s.svc.Cancel(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &orderpb.CancelResponse{
		Success: true,
	}, nil
}

func (s *Server) Get(ctx context.Context, req *orderpb.GetRequest) (*orderpb.GetResponse, error) {
	order, err := s.svc.Get(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &orderpb.GetResponse{
		Id:        order.ID,
		OrderNo:   order.OrderNo,
		UserId:    order.UserID,
		ProductId: order.ProductID,
		Count:     int64(order.Count),
		Amount:    order.Amount,
		Status:    order.Status,
	}, nil
}

func (s *Server) Pay(ctx context.Context, req *orderpb.PayRequest) (*orderpb.PayResponse, error) {
	err := s.svc.Pay(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &orderpb.PayResponse{
		Success: true,
	}, nil
}
