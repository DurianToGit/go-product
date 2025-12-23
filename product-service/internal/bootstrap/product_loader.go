package bootstrap

import (
	"context"
	"fmt"
	"product-service/internal/Cache"
	"product-service/internal/domain"
	"product-service/internal/service"
	"sync"
	"time"
)

func LoadProductsAsync(ctx context.Context, productChan chan<- domain.Product) {
	// 模拟外部数据源
	products := []domain.Product{
		{ID: 1, Name: "iPhone 17 Pro", Price: 9999, Stock: 100},
		{ID: 2, Name: "MacBook Pro ", Price: 19999, Stock: 50},
		{ID: 3, Name: "AirPods Pro", Price: 999, Stock: 200},
	}

	go func() {
		defer close(productChan)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(500 * time.Millisecond)
				productChan <- p
			}
		}
	}()
}

func LoadProducts(src *service.ProductService) {
	var wg sync.WaitGroup
	c := cache.NewProductCache()
	productChan := make(chan domain.Product)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	LoadProductsAsync(ctx, productChan)
	defer cancel()
	for {
		select {
		case product, ok := <-productChan:
			if !ok {
				wg.Wait()
				fmt.Println("product preload done")
				return
			}
			wg.Add(1)
			go func(ctx context.Context, p domain.Product) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					fmt.Println("product preload timeout, skip remaining")
					return
				default:
					c.Set(p)
				}
			}(ctx, product)
		case <-ctx.Done():
			fmt.Println("product preload timeout, skip remaining")
			return
		}
	}
}
