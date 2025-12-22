package bootstrap

import (
	"fmt"
	"product-service/internal/Cache"
	"product-service/internal/domain"
	"product-service/internal/service"
	"sync"
	"time"
)

func LoadProductsAsync(productChan chan<- domain.Product) {
	// 模拟外部数据源
	products := []domain.Product{
		{ID: 1, Name: "iPhone 17 Pro", Price: 9999, Stock: 100},
		{ID: 2, Name: "MacBook Pro ", Price: 19999, Stock: 50},
		{ID: 3, Name: "AirPods Pro", Price: 999, Stock: 200},
	}

	go func() {
		for _, p := range products {
			// time.Sleep(500 * time.Millisecond) // 模拟 IO
			// if p.Stock == 100 {
			// 	time.Sleep(2 * time.Second)
			// }
			productChan <- p
		}
		close(productChan) // 非常重要
	}()
}

func LoadProducts(src *service.ProductService) {
	var wg sync.WaitGroup
	cache := Cache.NewProductCache()
	productChan := make(chan domain.Product)
	LoadProductsAsync(productChan)
	timeout := time.After(time.Second * 3)
	// timeout := time.NewTimer(time.Second * 3)
	// defer timeout.Stop()
	for {
		select {
		case product, ok := <-productChan:
			if !ok {
				wg.Wait()
				fmt.Println("product preload done")
				for i, v := range cache.List() {
					fmt.Println("从缓存获取的")
					fmt.Println(i, v.Name)
				}
				return
			}
			wg.Add(1)
			go func(p domain.Product) {
				defer wg.Done()
				cache.Set(p)
				// err := src.CreateProduct(&product)
				// if err != nil {
				// 	log.Println("Error: ", err)
				// }
			}(product)
			// log.Println("Load product: ", product.Name)
		case <-timeout:
			fmt.Println("product preload timeout, skip remaining")
			return
		}
	}
}
