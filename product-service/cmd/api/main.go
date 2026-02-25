package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/config"
	"product-service/internal/logger"
	"product-service/internal/middleware"
	otelx "product-service/internal/otel"
	"product-service/internal/registry"
	"product-service/internal/router"
	"syscall"
	"time"
)

func main() {
	logger.InitFromEnv("product-service")
	defer logger.Sync()

	app := bootstrap.InitApp()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown-host"
	}
	pid := os.Getpid()
	inst := registry.ServiceInstance{
		ID:   fmt.Sprintf("%s-%d", hostName, pid), // 先写死，后面 D38 会改成 uuid/hostname+pid
		Addr: app.Config.App.Addr,                 // 你的实际监听地址
	}

	if app.EtcdLoader != nil {
		// 注册服务
		rerr := app.EtcdLoader.Register(ctx, "product-service", inst, 10)
		if rerr != nil {
			logger.L().Warn("registry_register_failed", zap.Error(rerr))
		}
	}

	// 创建Gin引擎实例
	r := gin.New()
	// 注册中间件：日志、耗时统计、异常恢复
	//r.Use(
	//	middleware.Logger(),   // 日志中间件：记录请求日志
	//	middleware.Cost(),     // 耗时中间件：统计请求处理时间
	//	middleware.Recovery(), // 恢复中间件：捕获panic并恢复
	//)
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(),
		middleware.RecoveryZap(),
	)

	// 注册路由：将所有API路由注册到引擎
	router.Register(r, app)

	// HTTP服务器也需要优雅关闭
	srv := &http.Server{
		Addr:    app.Config.App.Addr,
		Handler: r,
	}
	logger.L().Info("server_listening", zap.String("addr", app.Config.App.Addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("server_listen_failed", zap.Error(err))
		}
	}()

	shutdown, err := otelx.Init("product-service")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	<-ctx.Done()
	if app.EtcdLoader != nil {
		defer func(reg *config.EtcdLoader) {
			err = reg.Close()
			if err != nil {
				logger.L().Fatal("[registry] Close failed: %v\n", zap.Error(err))
			}
		}(app.EtcdLoader)
	}

	logger.L().Info("Shutting down server...")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	logger.L().Info("server_shutting_down")
	if err = srv.Shutdown(ctx2); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
