package bootstrap

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"product-service/internal/client/productclient"
	"product-service/internal/config"
	"product-service/internal/gateway/productgateway"
	"product-service/internal/validator"
	"product-service/pkg/breaker"
	"product-service/pkg/configwatch"
	"product-service/pkg/db"
	"product-service/pkg/grpcx"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	handlerOrder "product-service/services/order/handler"
	mysqlOrder "product-service/services/order/repository/mysql"
	modelOrder "product-service/services/order/repository/mysql/model"
	serviceOrder "product-service/services/order/service"
	handlerProduct "product-service/services/product/handler"
	mysqlProduct "product-service/services/product/repository/mysql"
	modelProduct "product-service/services/product/repository/mysql/model"
	serviceProduct "product-service/services/product/service"
	handlerUser "product-service/services/user/handler"
	mysqlUser "product-service/services/user/repository/mysql"
	modelUser "product-service/services/user/repository/mysql/model"
	serviceUser "product-service/services/user/service"
	"time"

	"github.com/joho/godotenv"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

type App struct {
	Config                *config.Config
	UserHandler           *handlerUser.UserHandler
	ProductHandler        *handlerProduct.ProductHandler
	ProductGatewayHandler *productgateway.Handler
	OrderHandler          *handlerOrder.OrderHandler

	MysqlClient  *gorm.DB
	RedisClient  *redis2.Client
	RedisBreaker *breaker.CircuitBreaker

	EtcdLoader  *config.EtcdLoader
	grpcConn    *grpc.ClientConn   // 存储 gRPC 连接以便关闭
	cancelWatch context.CancelFunc // 用于取消 etcd watch
}

// Close 优雅关闭应用资源
func (a *App) Close() error {
	var errs []error

	// 取消 etcd watch
	if a.cancelWatch != nil {
		a.cancelWatch()
	}

	// 关闭 gRPC 连接
	if a.grpcConn != nil {
		if err := a.grpcConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭 MySQL 连接
	if a.MysqlClient != nil {
		if sqlDB, err := a.MysqlClient.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 关闭 Redis 连接
	if a.RedisClient != nil {
		if err := a.RedisClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭 EtcdLoader
	if a.EtcdLoader != nil {
		if err := a.EtcdLoader.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 返回第一个错误（如果有）
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func BaseInit() *config.Config {
	// 加载环境变量
	_ = godotenv.Load()
	// 初始化验证器
	validator.Init()
	// 初始化随机数 不然rand的值会重复
	rand.Seed(time.Now().UnixNano())
	// 加载配置
	cfg := config.Load()

	return cfg
}

func InitApp() (*App, error) {
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

	err := configwatch.Watch(".env", func() {
		if err := godotenv.Overload(".env"); err != nil {
			logger.L().Error("reload .env failed", zap.Error(err))
			return
		}
		config.Reload()
		logger.L().Info("config reloaded", zap.Any("config", config.Get()))
	})
	if err != nil {
		logger.L().Error("加载环境变量失败", zap.Error(err))
		return nil, err
	}

	// 初始化 gRPC 连接
	var grpcConn *grpc.ClientConn
	if cfg.App.Product.Grpc {
		var grpcErr error
		// 使用 WithBlock() 阻塞直到连接建立，WithTimeout() 设置超时
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		grpcConn, grpcErr = grpc.DialContext(
			ctx,
			cfg.App.Product.GrpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		)
		if grpcErr != nil {
			logger.L().Error("初始化 grpc client 失败", zap.Error(grpcErr))
			return nil, grpcErr
		}
		cleanup = append(cleanup, func() {
			if err := grpcConn.Close(); err != nil {
				logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
			}
		})
	}

	// etcd 加载器
	var cancelWatch context.CancelFunc
	loader, err := config.NewEtcdLoader(cfg.Etcd.Endpoints)
	logger.L().Info("初始化etcd 加载器")
	if err != nil {
		logger.L().Error("初始化失败", zap.Error(err))
		config.SetRuntimeConfig(config.DefaultRuntimeConfig())
	} else {
		logger.L().Info("初始化etcd 加载器成功")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		// 启动时加载
		if err := loader.LoadOnce(ctx); err != nil {
			log.Printf("load config failed: %v", err)
			config.SetRuntimeConfig(config.DefaultRuntimeConfig())
		} else {
			logger.L().Info("load config success", zap.Any("config", config.GetRuntimeConfig()))
		}
		cancel()

		// 后台 etcd watch - 注意：不能使用 defer cancel2()，否则函数返回时 watch 会立即停止
		ctx2, cancel2 := context.WithCancel(context.Background())
		cancelWatch = cancel2
		go loader.Watch(ctx2)

		cleanup = append(cleanup, func() {
			cancelWatch()
			if err := loader.Close(); err != nil {
				logger.L().Error("关闭 EtcdLoader 失败", zap.Error(err))
			}
		})
	}

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

	// 熔断器
	redisBreaker := breaker.New(5, 10*time.Second) // 连续失败 5 次, 熔断器开启 10 秒
	redis.SetBreaker(redisBreaker)
	// 初始化kafka client
	kafka.InitClient(cfg.Kafka.Addr)

	// 迁移
	_ = mySQL.AutoMigrate(
		&modelProduct.ProductModel{},
		&modelProduct.ProductEventConsumedModel{},
		&modelOrder.OrderModel{},
		&modelOrder.OutboxEventModel{},
		&modelOrder.EventConsumeLog{},
		&modelUser.UserModel{},
	)

	// 初始化用户服务
	userRepo := mysqlUser.NewUserRepository(mySQL)
	userService := serviceUser.NewUserService(userRepo)
	userHandler := handlerUser.NewUserHandler(userService)

	// 初始化商品服务
	productRepo := mysqlProduct.NewProductRepository(mySQL)
	productService := serviceProduct.NewProductService(productRepo)
	productHandler := handlerProduct.NewProductHandler(productService)
	var productClient productclient.Client
	if cfg.App.Product.Grpc {
		productClient = productclient.NewGRPCClient(grpcConn)
	} else {
		productClient = productclient.NewLocalClient(productService)
	}
	productGatewayHandler := productgateway.NewHandler(productClient)

	// 初始化订单服务
	orderRepo := mysqlOrder.NewOrderRepository(mySQL)
	outboxRepo := mysqlOrder.NewOutboxRepository(mySQL)
	orderService := serviceOrder.NewOrderService(orderRepo, productClient, outboxRepo, mySQL)
	orderHandler := handlerOrder.NewOrderHandler(orderService)

	// 初始化成功，清空 cleanup 列表，防止 defer 执行清理
	cleanup = nil

	return &App{
		Config:                cfg,
		UserHandler:           userHandler,
		ProductHandler:        productHandler,
		ProductGatewayHandler: productGatewayHandler,
		OrderHandler:          orderHandler,
		MysqlClient:           mySQL,
		RedisClient:           rdb,
		RedisBreaker:          redisBreaker,
		EtcdLoader:            loader,
		grpcConn:              grpcConn,
		cancelWatch:           cancelWatch,
	}, nil
}
