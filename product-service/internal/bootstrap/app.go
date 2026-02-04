package bootstrap

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	redis2 "github.com/redis/go-redis/v9"
	"log"
	"math/rand"
	"product-service/api/handler"
	"product-service/internal/config"
	"product-service/internal/repository/mysql"
	"product-service/internal/repository/mysql/model"
	"product-service/internal/service"
	"product-service/internal/validator"
	"product-service/pkg/breaker"
	"product-service/pkg/db"
	"product-service/pkg/redis"
	"time"
)

type App struct {
	Config         *config.Config
	UserHandler    *handler.UserHandler
	ProductHandler *handler.ProductHandler
	OrderHandler   *handler.OrderHandler

	RedisClient  *redis2.Client
	RedisBreaker *breaker.CircuitBreaker
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
	// etcd 加载器
	loader, err := config.NewEtcdLoader(cfg.Etcd.Endpoints)
	log.Println("初始化etcd 加载器")
	if err != nil {
		log.Println("初始化失败")
		log.Printf("init etcd failed: %v", err)
		config.SetRuntimeConfig(config.DefaultRuntimeConfig())
	} else {
		log.Println("初始化成功")
		log.Println(loader)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// 启动时加载
		if err := loader.LoadOnce(ctx); err != nil {
			log.Printf("load config failed: %v", err)
			config.SetRuntimeConfig(config.DefaultRuntimeConfig())
		} else {
			log.Println("load config success")
		}
		log.Println(1233333)
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()

		// 后台 etcd watch
		go loader.Watch(ctx2)
	}

	return cfg
}

func InitApp() *App {
	cfg := BaseInit()
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
		&model.ProductModel{},
		&model.ProductEventConsumedModel{},
		&model.OrderModel{},
	)

	// 初始化用户服务
	userRepo := mysql.NewUserRepository(mySQL)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 初始化商品服务
	productRepo := mysql.NewProductRepository(mySQL)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// 初始化订单服务
	orderRepo := mysql.NewOrderRepository(mySQL)
	orderService := service.NewOrderService(orderRepo, productService)
	orderHandler := handler.NewOrderHandler(orderService)

	return &App{
		Config:         cfg,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
		OrderHandler:   orderHandler,
		RedisClient:    rdb,
		RedisBreaker:   redisBreaker,
	}
}
