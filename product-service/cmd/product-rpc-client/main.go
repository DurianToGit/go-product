package main

import (
	"context"
	"time"

	"product-service/internal/bootstrap"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"

	"go.uber.org/zap"

	"product-service/pkg/pb/productpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// 它做两件事：
//
// A. 调 gRPC health check
//
// 验证 product-rpc 是 SERVING
//
// B. 调 WatchProductStock
//
// 持续读 5 条消息打印出来

func main() {
	logger.InitFromEnv("product-client-rpc")
	defer logger.Sync()
	cfg := bootstrap.BaseInit()
	// 连接到 gRPC 服务端
	conn, err := grpc.Dial(
		cfg.App.Product.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		grpc.WithStreamInterceptor(grpcx.StreamClientLoggingInterceptor()),
	)
	if err != nil {
		logger.L().Fatal("连接 gRPC 服务失败", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
		if err != nil {
			logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
		}
	}(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置 request_id
	requestID := "client-" + time.Now().Format("20060102-150405")
	ctx = context.WithValue(ctx, grpcx.ContextRequestIDKey, requestID)

	// A. 调 gRPC health check
	logger.L().Info("===== A. Health Check =====")
	healthClient := healthpb.NewHealthClient(conn)
	healthResp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		logger.L().Fatal("Health check 失败", zap.Error(err))
	}
	if healthResp.GetStatus() == healthpb.HealthCheckResponse_SERVING {
		logger.L().Info("服务状态:", zap.String("status", healthResp.GetStatus().String()))
	} else {
		logger.L().Fatal("服务状态异常:", zap.String("status", healthResp.GetStatus().String()))
	}

	// B. 调 WatchProductStock
	logger.L().Info("===== B. WatchProductStock =====")
	productClient := productpb.NewProductServiceClient(conn)

	stream, err := productClient.WatchProductStock(ctx, &productpb.WatchProductStockRequest{
		ProductId: 1, // 监控商品 ID 为 1 的库存
	})
	if err != nil {
		logger.L().Fatal("调用 WatchProductStock 失败", zap.Error(err))
	}

	// 持续读 5 条消息
	for i := 0; i < 5; i++ {
		resp, err := stream.Recv()
		if err != nil {
			logger.L().Fatal("接收消息失败", zap.Error(err))
		}
		logger.L().Info("接收消息成功", zap.Int("index", i+1), zap.Int64("product_id", resp.GetProductId()), zap.Int64("stock", resp.GetStock()))
	}

	logger.L().Info("完成!")
}
