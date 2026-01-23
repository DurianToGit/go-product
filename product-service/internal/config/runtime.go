package config

import (
	"sync/atomic"
)

type RuntimeConfig struct {
	OrderCancelTimeoutSec int64
	ProductStreamKey      string
	ProductGroup          string
}

var runtimeCfg atomic.Value

func SetRuntimeConfig(cfg *RuntimeConfig) {
	runtimeCfg.Store(cfg)
}

func GetRuntimeConfig() *RuntimeConfig {
	cfg := runtimeCfg.Load()
	if cfg == nil {
		return nil
	}
	return cfg.(*RuntimeConfig)
}
