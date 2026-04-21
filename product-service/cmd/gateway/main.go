package main

import (
	"os"
	"os/signal"
	"syscall"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"
)

func main() {
	logger.InitFromEnv("gateway")
	defer logger.Sync()

	app, err := bootstrap.InitGatewayApp()
	if err != nil {
		logger.L().Fatal("init gateway app failed")
	}
	defer func(app *bootstrap.GatewayApp) {
		err = app.Close()
		if err != nil {
			logger.L().Error("gateway close failed")
		}
	}(app)

	logger.L().Info("gateway listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("gateway serve failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down gateway")
}
