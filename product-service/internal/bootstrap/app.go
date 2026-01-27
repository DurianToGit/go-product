package bootstrap

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"product-service/api/handler"
	"product-service/internal/config"
	"product-service/internal/repository/mysql"
	"product-service/internal/repository/mysql/model"
	"product-service/internal/service"
	"product-service/internal/validator"
	"product-service/pkg/db"
	"product-service/pkg/redis"
	"time"
)

type App struct {
	Config         *config.Config
	UserHandler    *handler.UserHandler
	ProductHandler *handler.ProductHandler
	OrderHandler   *handler.OrderHandler
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
	if err != nil {
		log.Fatalf("init etcd failed: %v", err)
	}

	ctx := context.Background()
	// 启动时加载
	if err := loader.LoadOnce(ctx); err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 后台 etcd watch
	go loader.Watch(ctx)

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
	redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	// 迁移
	_ = mySQL.AutoMigrate(
		&model.ProductModel{},
		&model.ProductEventConsumedModel{},
		&model.OrderModel{},
	)

	// etcdInit(&cfg.Etcd)

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
	}
}

func etcdInit(etcdConfig *config.EtcdConfig) {

}
