package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/repository/mysql"
	"product-service/internal/service"
	"product-service/pkg/db"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"syscall"
	"time"
)

func main() {
	logger.InitFromEnv("product-worker-cron")
	defer logger.Sync()
	cfg := bootstrap.BaseInit()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.DBUser,
		cfg.DB.DBPass,
		cfg.DB.DBHost,
		cfg.DB.DBPort,
		cfg.DB.DBName,
	)
	mySQL := db.InitMySQL(dsn)
	redis.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	// 初始化商品服务
	productRepo := mysql.NewProductRepository(mySQL)
	productService := service.NewProductService(productRepo)

	// 初始化订单服务
	orderRepo := mysql.NewOrderRepository(mySQL)
	orderService := service.NewOrderService(orderRepo, productService)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Println("取消订单进程Worker 启动成功，开始定时检测")
	for {
		select {
		case <-ctx.Done():
			log.Println("取消订单进程Worker 停止")
			return
		case <-ticker.C:
			n, err := orderService.CancelExpired(ctx, time.Now(), time.Minute*15)
			if err != nil {
				log.Printf("取消过期订单失败: %v", err)
			} else {
				log.Printf("取消过期订单数: %d", n)
			}
		}
	}
}
