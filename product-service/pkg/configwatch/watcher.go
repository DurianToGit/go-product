package configwatch

import (
	"go.uber.org/zap"
	"product-service/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

func Watch(file string, reload func()) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(file)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {

			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.L().Info("config file modified")
					reload()
				}

			case err := <-watcher.Errors:
				logger.L().Error("watch error", zap.Error(err))
			}
		}
	}()

	return nil
}
