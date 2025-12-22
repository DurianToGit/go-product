package app

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"product-service/api"
	"product-service/internal/bootstrap"
	"product-service/internal/config"
	"product-service/internal/database"
	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/internal/service"
)

func Run() {
	_ = godotenv.Load()

	cfg := config.Load()
	db := database.InitMySQL(cfg)

	productRepo := repository.NewMysqlProductRepo(db)
	productService := service.NewProductService(productRepo)
	// 加载数据
	if os.Getenv("LOAD_DATA") == "1" {
		bootstrap.LoadProducts(productService)
	}

	productHandler := handler.NewProductHandler(productService)

	r := api.NewRouter(productHandler)

	port := os.Getenv("PORT")
	log.Printf("Product Service running on :{%d}...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
