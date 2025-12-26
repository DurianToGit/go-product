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
	ProductHandler *handler.ProductHandler
	Config         *config.Config
}

func InitApp() *App {
	// 加载环境变量
	_ = godotenv.Load()
	// 初始化验证器
	validator.Init()
	// 加载配置
	cfg := config.Load()
	mySQL := db.InitMySQL(cfg.DB)
	productRepo := mysql.NewProductRepository(mySQL)

	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	return &App{
		ProductHandler: productHandler,
		Config:         cfg,
	}
}
