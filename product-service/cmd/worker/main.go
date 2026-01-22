package main

import (
	"context"
	"fmt"
	redis2 "github.com/redis/go-redis/v9"
	"log"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/domain"
	"product-service/internal/repository/mysql"
	"product-service/pkg/db"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/pkg/stream"
	"strconv"
	"syscall"
)

// 单独运行worker执行stream消费者逻辑

func main() {
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
	consumer := stream.NewProductEventConsumer(redis.Client, "product-consumer")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err := consumer.InitGroup(ctx)
	if err != nil {
		log.Println("init group error:", err)
		return
	}

	// 初始化商品服务
	productRepo := mysql.NewProductRepository(mySQL)
	go consumer.Consume(ctx, func(ctx context.Context, msg redis2.XMessage) error {
		// 解析 values
		eventType, _ := msg.Values["event_type"].(string)
		if eventType == domain.ProductEventTypeStockDeducted {
			return stockStockDeducted(ctx, msg, productRepo)
		} else if eventType == domain.ProductEventTypeRestockDeducted {
			return restockStockDeducted(ctx, msg, productRepo)
		}
		return nil
	})

	// 等待退出信号
	log.Println("Worker 启动成功，开始监听 stream...")
	<-ctx.Done()
	log.Println("收到退出信号，正在关闭...")
}

// 商品库存扣减
func stockStockDeducted(ctx context.Context, msg redis2.XMessage, productRepo *mysql.ProductRepository) error {
	productID, err := toInt64(msg.Values["product_id"])
	if err != nil {
		return err
	}
	count, err := toInt64(msg.Values["count"])
	if err != nil {
		return err
	}
	userId, err := toInt64(msg.Values["user_id"])
	if err != nil {
		return err
	}
	log.Printf("收到商品库存扣减事件：product_id=%d, count=%d, user_id=%d", productID, count, userId)
	// 幂等 + 扣库存（事务内）
	return productRepo.ConsumeStockDeductEvent(ctx, rediskeys.ProductStreamKey, msg.ID, productID, count, domain.ProductEventTypeStockDeducted)
}

// 恢复库存
func restockStockDeducted(ctx context.Context, msg redis2.XMessage, productRepo *mysql.ProductRepository) error {
	productID, err := toInt64(msg.Values["product_id"])
	if err != nil {
		return err
	}
	count, err := toInt64(msg.Values["count"])
	if err != nil {
		return err
	}
	log.Printf("收到商品库存恢复扣减事件：product_id=%d, count=%d", productID, count)
	// 幂等 + 扣库存（事务内）
	return productRepo.ConsumeRestockDeductEvent(ctx, rediskeys.ProductStreamKey, msg.ID, productID, count, domain.ProductEventTypeRestockDeducted)
}

func toInt64(v any) (int64, error) {
	switch t := v.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case string:
		return strconv.ParseInt(t, 10, 64)
	case []byte:
		return strconv.ParseInt(string(t), 10, 64)
	default:
		return 0, fmt.Errorf("invalid type: %T", v)
	}
}
