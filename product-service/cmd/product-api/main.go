package main

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"
)

func main() {
	logger.InitFromEnv("product-api")
	defer logger.Sync()

	app, err := bootstrap.InitProductApp()
	if err != nil {
		logger.L().Fatal("init product app failed")
	}
	defer func(app *bootstrap.ProductApp) {
		err = app.Close()
		if err != nil {
			logger.L().Error("close product app failed", zap.Error(err))
		}
	}(app)

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
}
