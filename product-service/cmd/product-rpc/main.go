package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"product-service/pkg/grpcx"
	"syscall"
	"time"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"product-service/internal/bootstrap"
	"product-service/internal/config"
	"product-service/pkg/breaker"
	"product-service/pkg/configwatch"
	"product-service/pkg/db"
	"product-service/pkg/logger"
	"product-service/pkg/pb/productpb"
	"product-service/pkg/redis"
	serviceProductRpc "product-service/services/product/grpc"
	mysqlProduct "product-service/services/product/repository/mysql"
	serviceProduct "product-service/services/product/service"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.InitFromEnv("product-rpc")
	defer logger.Sync()
	cfg := bootstrap.BaseInit()
	err := configwatch.Watch(".env", func() {
		if err := godotenv.Overload(".env"); err != nil {
			logger.L().Error("reload .env failed", zap.Error(err))
			return
		}
		config.Reload()
		logger.L().Info("config reloaded", zap.Any("config", config.Get()))
	})
	if err != nil {
		logger.L().Fatal("加载环境变量失败", zap.Error(err))
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	if err != nil {
		logger.L().Fatal("初始化数据库失败", zap.Error(err))
	}
	_ = redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	// 熔断器
	redisBreaker := breaker.New(5, 10*time.Second) // 连续失败 5 次, 熔断器开启 10 秒
	redis.SetBreaker(redisBreaker)

	// 初始化商品服务
	productRepo := mysqlProduct.NewProductRepository(mySQL)
	productService := serviceProduct.NewProductService(productRepo)
	rpcProductService := serviceProductRpc.NewServer(productService)
	lis, err := net.Listen("tcp", cfg.App.Product.GrpcAddr)
	if err != nil {
		logger.L().Fatal("监听端口失败", zap.Error(err))
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcx.UnaryServerRecoveryInterceptor(),
			grpcx.UnaryServerMetadataInterceptor(),
			grpcx.UnaryServerLoggingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpcx.StreamServerMetadataInterceptor(),
			grpcx.StreamServerLoggingInterceptor(),
		),
	)
	productpb.RegisterProductServiceServer(grpcServer, rpcProductService)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// 优雅关闭处理
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.L().Info("接收到关闭信号，正在优雅关闭服务...")
		grpcServer.GracefulStop()
	}()

	logger.L().Info("gRPC 服务启动成功", zap.String("addr", cfg.App.Product.GrpcAddr))
	if err := grpcServer.Serve(lis); err != nil {
		logger.L().Fatal("启动服务失败", zap.Error(err))
	}
}
