package main

import (
	"os"
	"os/signal"
	"syscall"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"
)

func main() {
	logger.InitFromEnv("order-api")
	defer logger.Sync()

	app, err := bootstrap.InitOrderApp()
	if err != nil {
		logger.L().Fatal("init order app failed")
	}
	defer func(app *bootstrap.OrderApp) {
		err = app.Close()
		if err != nil {
			logger.L().Error("order app close failed")
		}
	}(app)

	logger.L().Info("order-api listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("order api serve failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down order-api")
}
