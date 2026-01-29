package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"product-service/internal/bootstrap"
	"product-service/internal/middleware"
	otelx "product-service/internal/otel"
	"product-service/internal/registry"
	"product-service/internal/router"
	"syscall"
	"time"
)

func main() {

	app := bootstrap.InitApp()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// 初始化服务注册中心
	reg, err := registry.NewEtcdRegistry(app.Config.Etcd.Endpoints)
	if err != nil {
		log.Printf("[registry]  Failed to connect to etcd: %v\n", err)
		return
	}
	defer func(reg *registry.EtcdRegistry) {
		err := reg.Close()
		if err != nil {
			log.Printf("[registry] Close failed: %v\n", err)
		}
	}(reg)
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown-host"
	}
	pid := os.Getpid()
	inst := registry.ServiceInstance{
		ID:   fmt.Sprintf("%s-%d", hostName, pid), // 先写死，后面 D38 会改成 uuid/hostname+pid
		Addr: app.Config.App.Addr,                 // 你的实际监听地址
	}

	rerr := reg.Register(ctx, "product-service", inst, 10)
	if rerr != nil {
		log.Printf("[registry] Register failed: %v\n", rerr)
	}

	// 创建Gin引擎实例
	r := gin.New()
	// 注册中间件：日志、耗时统计、异常恢复
	r.Use(
		middleware.Logger(),   // 日志中间件：记录请求日志
		middleware.Cost(),     // 耗时中间件：统计请求处理时间
		middleware.Recovery(), // 恢复中间件：捕获panic并恢复
	)

	// 注册路由：将所有API路由注册到引擎
	router.Register(r, app)

	// HTTP服务器也需要优雅关闭
	srv := &http.Server{
		Addr:    app.Config.App.Addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	shutdown, err := otelx.Init("product-service")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
