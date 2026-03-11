package bootstrap

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"product-service/api/handler"
	"product-service/internal/config"
	"product-service/internal/repository/mysql"
	"product-service/internal/service"
	"product-service/internal/validator"
	"product-service/pkg/breaker"
	"product-service/pkg/configwatch"
	"product-service/pkg/db"
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
	"time"
)

type App struct {
	Config         *config.Config
	UserHandler    *handler.UserHandler
	ProductHandler *handlerProduct.ProductHandler
	OrderHandler   *handlerOrder.OrderHandler

	MysqlClient  *gorm.DB
	RedisClient  *redis2.Client
	RedisBreaker *breaker.CircuitBreaker

	EtcdLoader *config.EtcdLoader
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

func InitApp() *App {
	cfg := BaseInit()
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
		return nil
	}
	// etcd 加载器
	loader, err := config.NewEtcdLoader(cfg.Etcd.Endpoints)
	logger.L().Info("初始化etcd 加载器")
	if err != nil {
		logger.L().Error("初始化失败", zap.Error(err))
		config.SetRuntimeConfig(config.DefaultRuntimeConfig())
	} else {
		logger.L().Info("初始化etcd 加载器成功")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// 启动时加载
		if err := loader.LoadOnce(ctx); err != nil {
			log.Printf("load config failed: %v", err)
			config.SetRuntimeConfig(config.DefaultRuntimeConfig())
		} else {
			logger.L().Info("load config success", zap.Any("config", config.GetRuntimeConfig()))
		}
		ctx2, _ := context.WithCancel(context.Background())

		// 后台 etcd watch
		go loader.Watch(ctx2)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	rdb := redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	// 熔断器
	redisBreaker := breaker.New(5, 10*time.Second) // // 连续失败 5 次, 熔断器开启 10 秒
	redis.SetBreaker(redisBreaker)

	// 迁移
	_ = mySQL.AutoMigrate(
		&modelProduct.ProductModel{},
		&modelProduct.ProductEventConsumedModel{},
		&modelOrder.OrderModel{},
	)

	// 初始化用户服务
	userRepo := mysql.NewUserRepository(mySQL)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 初始化商品服务
	productRepo := mysqlProduct.NewProductRepository(mySQL)
	productService := serviceProduct.NewProductService(productRepo)
	productHandler := handlerProduct.NewProductHandler(productService)

	// 初始化订单服务
	orderRepo := mysqlOrder.NewOrderRepository(mySQL)
	orderService := serviceOrder.NewOrderService(orderRepo)
	orderHandler := handlerOrder.NewOrderHandler(orderService)

	return &App{
		Config:         cfg,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
		OrderHandler:   orderHandler,
		MysqlClient:    mySQL,
		RedisClient:    rdb,
		RedisBreaker:   redisBreaker,
		EtcdLoader:     loader,
	}
}
