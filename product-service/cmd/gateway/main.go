package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	logger.InitFromEnv("gateway")
	defer logger.Sync()

	app, err := bootstrap.InitGatewayApp()
	if err != nil {
		logger.L().Fatal("init gateway app failed", zap.Error(err))
	}

	logger.L().Info("gateway listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("gateway serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down gateway")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	done := make(chan struct{})
	go func() {
		if err := app.Close(); err != nil {
			logger.L().Error("gateway close failed", zap.Error(err))
		}
		close(done)
	}()
	select {
	case <-done:
		logger.L().Info("gateway stopped gracefully")
	case <-shutdownCtx.Done():
		logger.L().Error("gateway shutdown timed out", zap.Error(shutdownCtx.Err()))
	}
}
