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
	logger.InitFromEnv("order-api")
	defer logger.Sync()

	app, err := bootstrap.InitOrderApp()
	if err != nil {
		logger.L().Fatal("init order app failed", zap.Error(err))
	}

	logger.L().Info("order-api listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("order api serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down order-api")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := app.Close(); err != nil {
		logger.L().Error("order app close failed", zap.Error(err))
	}
	logger.L().Info("order-api stopped", zap.Error(shutdownCtx.Err()))
}
