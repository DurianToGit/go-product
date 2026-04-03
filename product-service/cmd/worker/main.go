package main

import (
	"context"
	"fmt"
	redis2 "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"log"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/config"
	"product-service/internal/registry"
	"product-service/pkg/db"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/pkg/stream"
	"product-service/services/order/domain"
	"product-service/services/product/repository/mysql"
	"strconv"
	"syscall"
)

// 单独运行worker执行stream消费者逻辑

func main() {
	logger.InitFromEnv("product-worker")
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
	consumer := stream.NewProductEventConsumer(redis.Client, "product-consumer")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err := consumer.InitGroup(ctx)
	if err != nil {
		logger.L().Error("init group error", zap.Error(err))
		return
	}

	// 初始化服务注册中心
	reg, err := config.NewEtcdLoader(cfg.Etcd.Endpoints)
	if err != nil {
		log.Printf("初始化注册中心失败：%v", err)
	} else {
		inst := registry.ServiceInstance{
			ID:   "worker-1",    // 先写死，后面 D38 会改成 uuid/hostname+pid
			Addr: "127.0.0.1:0", // 你的实际监听地址
		}

		rerr := reg.Register(ctx, "product-worker", inst, 10)
		if rerr != nil {
			log.Printf("[registry] Register failed: %v", rerr)
		}
	}

	/*shutdown, err := otelx.Init("product-worker")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()*/

	// 初始化商品服务
	productRepo := mysql.NewProductRepository(mySQL)
	go consumer.Consume(ctx, func(ctx context.Context, msg redis2.XMessage) error {
		// ctx2 := extractCtx(ctx, msg.Values)
		// ctx2, span := otel.Tracer("worker").Start(ctx2, "stream.consume")
		// defer span.End()
		// 解析 values
		eventType, _ := msg.Values["event_type"].(string)
		logger.L().Info("收到商品事件", zap.String("event_type", eventType))
		if eventType == domain.ProductEventTypeStockDeducted {
			return stockStockDeducted(ctx, msg, productRepo)
		} else if eventType == domain.ProductEventTypeRestockDeducted {
			return restockStockDeducted(ctx, msg, productRepo)
		}
		logger.L().Warn("未知商品事件", zap.String("event_type", eventType))
		return nil
	})

	// 等待退出信号
	logger.L().Info("Worker 启动成功，开始监听 stream...")
	<-ctx.Done()
	if reg != nil {
		defer func(reg *config.EtcdLoader) {
			err := reg.Close()
			if err != nil {

			}
		}(reg)
	}

	logger.L().Info("收到退出信号，正在关闭...")
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
	log.Printf("收到恢复商品库存事件：product_id=%d, count=%d", productID, count)
	// 幂等 + 扣库存（事务内）
	return productRepo.ConsumeRestockDeductEvent(ctx, rediskeys.ProductStreamKey, msg.ID, productID, count, domain.ProductEventTypeRestockDeducted)
}

func extractCtx(ctx context.Context, values map[string]any) context.Context {
	tp, _ := values["traceparent"].(string)
	if tp == "" {
		return ctx
	}
	carrier := propagation.MapCarrier{
		"traceparent": tp,
	}
	// tracestate 可选
	if ts, ok := values["tracestate"].(string); ok && ts != "" {
		carrier["tracestate"] = ts
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
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
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("invalid type: %T", v)
	}
}
