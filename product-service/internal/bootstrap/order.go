package bootstrap

import (
	"context"
	"fmt"
	"net"
	"product-service/internal/client/productclient"
	"product-service/pkg/db"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/pb/orderpb"
	"product-service/pkg/redis"
	ordergrpc "product-service/services/order/grpc"
	orderRepository "product-service/services/order/repository"
	orderMysql "product-service/services/order/repository/mysql"
	modelOrder "product-service/services/order/repository/mysql/model"
	orderService "product-service/services/order/service"
	"time"

	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

type OrderApp struct {
	grpcServer *grpc.Server
	listener   net.Listener

	mysqlDB *gorm.DB
	redis   *redis2.Client

	productConn   *grpc.ClientConn
	productClient productclient.Client

	orderRepo    orderRepository.OrderRepository
	outboxRepo   orderRepository.OutboxRepository
	orderService *orderService.OrderService
	orderRPC     *ordergrpc.Server
}

func InitOrderApp() (*OrderApp, error) {
	cfg := BaseInit()
	// 资源清理函数列表，用于初始化失败时回滚
	var cleanup []func()

	// 注册清理函数，如果后续初始化失败则执行
	defer func() {
		if cleanup != nil {
			// 初始化失败，执行所有清理函数
			for i := len(cleanup) - 1; i >= 0; i-- {
				cleanup[i]()
			}
		}
	}()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	cleanup = append(cleanup, func() {
		if sqlDB, err := mySQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.L().Error("关闭 MySQL 连接失败", zap.Error(err))
			}
		}
	})

	rdb := redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	cleanup = append(cleanup, func() {
		if err := rdb.Close(); err != nil {
			logger.L().Error("关闭 Redis 连接失败", zap.Error(err))
		}
	})

	// 初始化 gRPC 连接
	var productGrpcConn *grpc.ClientConn
	var grpcErr error
	// 使用 WithBlock() 阻塞直到连接建立，WithTimeout() 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	productGrpcConn, grpcErr = grpc.DialContext(
		ctx,
		cfg.App.Product.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
	)
	cancel()
	if grpcErr != nil {
		logger.L().Error("初始化 grpc client 失败", zap.Error(grpcErr))
		return nil, grpcErr
	}
	cleanup = append(cleanup, func() {
		if err := productGrpcConn.Close(); err != nil {
			logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
		}
	})

	_ = mySQL.AutoMigrate(
		&modelOrder.OrderModel{},
		&modelOrder.OutboxEventModel{},
		&modelOrder.EventConsumeLog{},
	)

	orderRepo := orderMysql.NewOrderRepository(mySQL)
	outboxRepo := orderMysql.NewOutboxRepository(mySQL)
	productClient := productclient.NewGRPCClient(productGrpcConn)
	orderService1 := orderService.NewOrderService(orderRepo, productClient, outboxRepo, mySQL)

	// productRepo := productMysql.NewProductRepository(mySQL)
	// productService := productService.NewProductService(productRepo)

	rpcOrderService := ordergrpc.NewServer(orderService1)
	// rpcProductService := productRPC.NewServer(productService)
	lis, err := net.Listen("tcp", cfg.App.Order.GrpcAddr)
	if err != nil {
		return nil, fmt.Errorf("监听端口失败: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcx.UnaryServerRecoveryInterceptor(),
			grpcx.UnaryServerMetadataInterceptor(),
			grpcx.UnaryServerLoggingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpcx.StreamServerRecoveryInterceptor(),
			grpcx.StreamServerMetadataInterceptor(),
			grpcx.StreamServerLoggingInterceptor(),
		),
	)
	orderpb.RegisterOrderServiceServer(grpcServer, rpcOrderService)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	cleanup = nil

	return &OrderApp{
		grpcServer:    grpcServer,
		listener:      lis,
		mysqlDB:       mySQL,
		redis:         rdb,
		productConn:   productGrpcConn,
		productClient: productClient,
		orderRepo:     orderRepo,
		outboxRepo:    outboxRepo,
		orderService:  orderService1,
		orderRPC:      rpcOrderService,
	}, nil
}

func (a *OrderApp) Serve() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *OrderApp) Close() error {
	var errs []error
	a.grpcServer.GracefulStop()
	// GracefulStop 已关闭 listener，无需再次关闭

	// 关闭 gRPC 客户端连接
	if a.productConn != nil {
		if err := a.productConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭 MySQL 连接
	if a.mysqlDB != nil {
		if sqlDB, err := a.mysqlDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 关闭 Redis 连接
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
