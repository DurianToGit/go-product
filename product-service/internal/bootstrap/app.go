package bootstrap

import (
	"github.com/joho/godotenv"
	"product-service/api/handler"
	"product-service/internal/config"
	"product-service/internal/repository/mysql"
	"product-service/internal/service"
	"product-service/internal/validator"
	"product-service/pkg/db"
)

type App struct {
	Config         *config.Config
	UserHandler    *handler.UserHandler
	ProductHandler *handler.ProductHandler
}

func InitApp() *App {
	// 加载环境变量
	_ = godotenv.Load()
	// 初始化验证器
	validator.Init()
	// 加载配置
	cfg := config.Load()
	mySQL := db.InitMySQL(cfg.DB)

	// 初始化用户服务
	userRepo := mysql.NewUserRepository(mySQL)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 初始化商品服务
	productRepo := mysql.NewProductRepository(mySQL)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	return &App{
		Config:         cfg,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
	}
}
