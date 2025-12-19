package bootstrap

import (
	"fmt"
	"log"
	"product-service/internal/domain"
	"product-service/internal/service"
	"time"
)

func LoadProductsAsync(productChan chan<- domain.Product) {
	// 模拟外部数据源
	products := []domain.Product{
		{Name: "iPhone 17 Pro", Price: 9999, Stock: 100},
		{Name: "MacBook Pro ", Price: 19999, Stock: 50},
		{Name: "AirPods Pro", Price: 999, Stock: 200},
	}

	go func() {
		for _, p := range products {
			time.Sleep(500 * time.Millisecond) // 模拟 IO
			if p.Stock == 100 {
				time.Sleep(2 * time.Second)
			}
			productChan <- p
		}
		close(productChan) // 非常重要
	}()
}

func LoadProducts(src *service.ProductService) {
	productChan := make(chan domain.Product)
	LoadProductsAsync(productChan)
	timeout := time.After(time.Second * 3)
	// timeout := time.NewTimer(time.Second * 3)
	// defer timeout.Stop()
	for {
		select {
		case product, ok := <-productChan:
			if !ok {
				fmt.Println("product preload done")
				return
			}
			log.Println("Load product: ", product.Name)
			err := src.CreateProduct(&product)
			if err != nil {
				log.Println("Error: ", err)
			}
		case <-timeout:
			fmt.Println("product preload timeout, skip remaining")
			return
		}
	}
}
