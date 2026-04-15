package grpc

import (
	"context"

	"product-service/pkg/pb/productpb"
	productService "product-service/services/product/service"
)

type Server struct {
	productpb.UnimplementedProductServiceServer
	svc *productService.ProductService
}

func NewServer(svc *productService.ProductService) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	p, err := s.svc.GetProduct(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &productpb.GetProductResponse{
		Id:    p.ID,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}, nil
}

func (s *Server) GetStock(ctx context.Context, req *productpb.GetStockRequest) (*productpb.GetStockResponse, error) {
	stock, err := s.svc.GetStock(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &productpb.GetStockResponse{
		Stock: stock,
	}, nil
}
