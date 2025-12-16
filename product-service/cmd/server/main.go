package main

import (
	"github.com/joho/godotenv"
	"log"
	"product-service/api"
	"product-service/internal/config"
	"product-service/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	// fmt.Println(cfg)

	db := database.InitMySQL(cfg)

	r := api.NewRouter(db)

	log.Println("Product Service running on :8080...")
	if err := r.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
