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
	logger.InitFromEnv("order-worker")
	defer logger.Sync()

	app, err := bootstrap.InitOrderWorkerApp()
	if err != nil {
		logger.L().Fatal("init order worker app failed", zap.Error(err))
	}

	logger.L().Info("order worker listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("order worker serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down order-worker")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	done := make(chan struct{})
	go func() {
		if err := app.Close(); err != nil {
			logger.L().Error("close order worker failed", zap.Error(err))
		}
		close(done)
	}()
	select {
	case <-done:
		logger.L().Info("order-worker stopped gracefully")
	case <-shutdownCtx.Done():
		logger.L().Error("order-worker shutdown timed out", zap.Error(shutdownCtx.Err()))
	}
}
