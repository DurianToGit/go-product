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
	logger.InitFromEnv("product-api")
	defer logger.Sync()

	app, err := bootstrap.InitProductApp()
	if err != nil {
		logger.L().Fatal("init product app failed", zap.Error(err))
	}

	logger.L().Info("product-api started")

	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("product api serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down product-api")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	done := make(chan struct{})
	go func() {
		if err := app.Close(); err != nil {
			logger.L().Error("close product app failed", zap.Error(err))
		}
		close(done)
	}()
	select {
	case <-done:
		logger.L().Info("product-api stopped gracefully")
	case <-shutdownCtx.Done():
		logger.L().Error("product-api shutdown timed out", zap.Error(shutdownCtx.Err()))
	}
}
