package bootstrap

import (
	"fmt"
	"net"
	"product-service/pkg/db"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/pb/productpb"
	"product-service/pkg/redis"
	productRPC "product-service/services/product/grpc"
	"product-service/services/product/repository"
	mysqlProduct "product-service/services/product/repository/mysql"
	modelProduct "product-service/services/product/repository/mysql/model"
	productService "product-service/services/product/service"

	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

type ProductApp struct {
	grpcServer *grpc.Server
	listener   net.Listener

	mysqlDB *gorm.DB
	redis   *redis2.Client

	productRepo    repository.ProductRepository
	productService *productService.ProductService
	productRPC     *productRPC.Server
}

func InitProductApp() (*ProductApp, error) {
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

	// 初始化商品服务
	productRepo := mysqlProduct.NewProductRepository(mySQL)
	service := productService.NewProductService(productRepo)
	rpcProductService := productRPC.NewServer(service)
	lis, err := net.Listen("tcp", cfg.App.Product.GrpcAddr)
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
	productpb.RegisterProductServiceServer(grpcServer, rpcProductService)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	_ = mySQL.AutoMigrate(
		&modelProduct.ProductModel{},
		&modelProduct.ProductEventConsumedModel{},
	)

	cleanup = nil
	return &ProductApp{
		grpcServer:     grpcServer,
		listener:       lis,
		mysqlDB:        mySQL,
		redis:          rdb,
		productRepo:    productRepo,
		productService: service,
		productRPC:     rpcProductService,
	}, nil
}

func (a *ProductApp) Serve() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *ProductApp) Close() error {
	var errs []error
	a.grpcServer.GracefulStop()
	// GracefulStop 已关闭 listener，无需再次关闭

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
