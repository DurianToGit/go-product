package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"product-service/internal/router"
)

func main() {
	r := gin.New()
	router.Register(r)
	if err := r.Run(":8082"); err != nil {
		log.Fatalln(err)
	}
}
