package app

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"product-service/internal/bootstrap"
	"product-service/internal/config"
	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/internal/router"
	"product-service/internal/service"
	db2 "product-service/pkg/db"
)

func Run() {
	_ = godotenv.Load()

	cfg := config.Load()
	db := db2.InitMySQL(cfg)

	productRepo := repository.NewMysqlProductRepo(db)
	productService := service.NewProductService(productRepo)
	// 加载数据
	if os.Getenv("LOAD_DATA") == "1" {
		bootstrap.LoadProducts(productService)
	}

	productHandler := handler.NewProductHandler(productService)

	r := router.NewRouter(productHandler)

	port := os.Getenv("PORT")
	log.Printf("Product Service running on :{%d}...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
