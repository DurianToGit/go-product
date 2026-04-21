package bootstrap

import (
	"fmt"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"net"
	"product-service/pkg/db"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/pb/userpb"
	"product-service/pkg/redis"
	usergrpc "product-service/services/user/grpc"
	userRepository "product-service/services/user/repository"
	userMysql "product-service/services/user/repository/mysql"
	userService "product-service/services/user/service"
)

type UserApp struct {
	grpcServer *grpc.Server
	listener   net.Listener

	mysqlDB *gorm.DB
	redis   *redis2.Client

	userRepo    userRepository.UserRepository
	userService *userService.UserService
	userRPC     *usergrpc.Server
}

func InitUserApp() (*UserApp, error) {
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
	userRepo := userMysql.NewUserRepository(mySQL)
	userService1 := userService.NewUserService(userRepo)
	userRPC := usergrpc.NewServer(userService1)

	lis, err := net.Listen("tcp", cfg.App.User.GrpcAddr)
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
	userpb.RegisterUserServiceServer(grpcServer, userRPC)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	cleanup = nil

	return &UserApp{
		grpcServer:  grpcServer,
		listener:    lis,
		mysqlDB:     mySQL,
		redis:       rdb,
		userRepo:    userRepo,
		userService: userService1,
		userRPC:     userRPC,
	}, nil

}

func (a *UserApp) Serve() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *UserApp) Close() error {
	var errs []error
	a.grpcServer.GracefulStop()
	// 关闭 MySQL 连接
	if a.mysqlDB != nil {
		if sqlDB, err := a.mysqlDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	err := a.redis.Close()
	if err != nil {
		errs = append(errs, err)
	}
	err = a.listener.Close()
	if err != nil {
		errs = append(errs, err)
	}
	return nil
}
