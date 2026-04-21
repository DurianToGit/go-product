package main

import (
	"os"
	"os/signal"
	"syscall"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"
)

func main() {
	logger.InitFromEnv("order-worker")
	defer logger.Sync()

	app, err := bootstrap.InitOrderWorkerApp()
	if err != nil {
		logger.L().Fatal("init order worker app failed")
	}
	defer func(app *bootstrap.OrderWorkerApp) {
		err = app.Close()
		if err != nil {
			logger.L().Error("close order worker failed")
		}
	}(app)

	logger.L().Info("order worker listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("order worker serve failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down order-worker")
}
