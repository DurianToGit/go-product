package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/client/productclient"
	"product-service/pkg/db"
	"product-service/pkg/grpcx"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/services/order/repository/mysql"
	"product-service/services/order/service"
	"syscall"
	"time"
)

// 订单生产者 使用kafka 创建生产者

func main() {
	logger.InitFromEnv("product-worker-cron")
	defer logger.Sync()
	cfg := bootstrap.BaseInit()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	// 初始化kafka client
	kafka.InitClient(cfg.Kafka.Addr)
	outboxRepository := mysql.NewOutboxRepository(mySQL)
	outboxRelay := service.NewOutboxRelay(outboxRepository)
	redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

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

	productClient := productclient.NewGRPCClient(conn)

	// 初始化订单服务
	repo := mysql.NewOrderRepository(mySQL)
	outboxRepo := mysql.NewOutboxRepository(mySQL)
	orderService := service.NewOrderService(repo, productClient, outboxRepo, mySQL)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	tickerOutBox := time.NewTicker(time.Minute)
	defer tickerOutBox.Stop()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.L().Info("取消订单进程Worker 启动成功，开始定时检测")
	for {
		select {
		case <-ctx.Done():
			logger.L().Info("取消订单进程Worker 停止")
			return
		case <-ticker.C:
			n, err := orderService.CancelExpired(ctx, time.Now(), time.Minute*15)
			if err != nil {
				logger.L().Error("取消过期订单失败", zap.Error(err))
			} else {
				logger.L().Info("取消过期订单数", zap.Int64("order_number", n))
			}
		case <-tickerOutBox.C:
			n, err := outboxRelay.RunOnce(ctx, 100)
			if err != nil {
				logger.L().Error("outbox执行失败", zap.Error(err))
			} else {
				logger.L().Info("outbox执行成功", zap.Int64("success_count", n))
			}

		}
	}
}
