package main

import (
	"os"
	"os/signal"
	"syscall"

	"product-service/internal/bootstrap"
	"product-service/pkg/logger"
)

func main() {
	logger.InitFromEnv("user-api")
	defer logger.Sync()

	app, err := bootstrap.InitUserApp()
	if err != nil {
		logger.L().Fatal("init user app failed")
	}
	defer func(app *bootstrap.UserApp) {
		err = app.Close()
		if err != nil {
			logger.L().Error("user app close failed")
		}
	}(app)

	logger.L().Info("user-api listening")
	go func() {
		if err := app.Serve(); err != nil {
			logger.L().Error("user api serve failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down user-api")
}
