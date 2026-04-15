package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	otelx "product-service/internal/otel"
	"product-service/internal/registry"
	"product-service/internal/router"
	"product-service/pkg/kafka"
	"product-service/pkg/logger"
	"product-service/pkg/middleware"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger.InitFromEnv("product-service")
	defer logger.Sync()

	app, err := bootstrap.InitApp()
	if err != nil {
		logger.L().Error("初始化应用失败", zap.Error(err))
		return
	}
	defer func(app *bootstrap.App) {
		err := app.Close()
		if err != nil {
			logger.L().Error("应用资源关闭失败", zap.Error(err))
			return
		}
	}(app)
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
	// r.Use(
	// 	middleware.Logger(),   // 日志中间件：记录请求日志
	// 	middleware.Recovery(), // 恢复中间件：捕获panic并恢复
	// )
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(),
		middleware.RecoveryZap(),
		middleware.Cost(), // 耗时中间件：统计请求处理时间
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
	if err != nil {
		logger.L().Error("otel_init_failed", zap.Error(err))
	}

	<-ctx.Done()
	logger.L().Info("接收到关闭信号，正在优雅关闭服务...")

	// 创建关闭超时上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 1. 关闭 HTTP 服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.L().Error("HTTP服务器关闭失败", zap.Error(err))
	}
	logger.L().Info("HTTP服务器已关闭")

	// 2. 关闭 Kafka 客户端
	if kafka.Client != nil {
		if err := kafka.Client.Close(); err != nil {
			logger.L().Error("Kafka关闭失败", zap.Error(err))
		}
		logger.L().Info("Kafka已关闭")
	}

	// 3. 关闭 OpenTelemetry
	if shutdown != nil {
		if err := shutdown(shutdownCtx); err != nil {
			logger.L().Error("OpenTelemetry关闭失败", zap.Error(err))
		}
	}

	logger.L().Info("应用资源已关闭")

	logger.L().Info("服务已优雅关闭")
}
