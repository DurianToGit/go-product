package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"product-service/pkg/grpcx"
	"product-service/pkg/logger"

	"go.uber.org/zap"

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
	rid := grpcx.GetRequestIDFromContext(ctx)
	logger.L().Info("grpc get product", zap.String("request_id", rid))
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

func (s *Server) WatchProductStock(req *productpb.WatchProductStockRequest, stream productpb.ProductService_WatchProductStockServer) error {
	// 不要先搞复杂 watch 机制，先做最小版：
	//
	// 从 req.ProductId 取商品 ID
	// 循环 5 次
	// 每次：
	// 查一次 s.svc.GetStock(...)
	// stream.Send(...)
	// time.Sleep(1 * time.Second)
	// 如果 stream.Context().Done() 了，立刻结束
	productID := req.ProductId
	ctx := stream.Context()

	rid := grpcx.GetRequestIDFromContext(ctx)
	logger.L().Info("Watch 产品库存",
		zap.Int64("product_id", productID),
		zap.String("request_id", rid),
	)

	for i := 0; i < 5; i++ {
		// 检查客户端是否已断开或等待 1 秒
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			// 查询库存
			stock, err := s.svc.GetStock(ctx, productID)
			if err != nil {
				return err
			}

			// 发送响应
			resp := &productpb.WatchProductStockResponse{
				ProductId: productID,
				Stock:     stock,
				Seq:       int64(i + 1),
			}
			if err = stream.Send(resp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) RestoreStock(ctx context.Context, req *productpb.RestoreStockRequest) (*productpb.RestoreStockResponse, error) {
	err := s.svc.RestoreStock(ctx, req.ProductId, req.Count, req.Source)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "restore stock failed: %v", err)
	}
	return &productpb.RestoreStockResponse{Success: true}, nil
}

func (s *Server) ConsumeStockDeductEvent(ctx context.Context, req *productpb.ConsumeStockDeductEventRequest) (*productpb.ConsumeStockDeductEventResponse, error) {
	err := s.svc.ConsumeStockDeductEvent(ctx, req.Stream, req.MsgId, req.ProductId, req.Count, req.EventType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "consume stock deduct event failed: %v", err)
	}
	return &productpb.ConsumeStockDeductEventResponse{Success: true}, nil
}

func (s *Server) ConsumeRestockDeductEvent(ctx context.Context, req *productpb.ConsumeStockDeductEventRequest) (*productpb.ConsumeStockDeductEventResponse, error) {
	err := s.svc.ConsumeRestockDeductEvent(ctx, req.Stream, req.MsgId, req.ProductId, req.Count, req.EventType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "consume restock deduct event failed: %v", err)
	}
	return &productpb.ConsumeStockDeductEventResponse{Success: true}, nil
}
