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

func DefaultRuntimeConfig() *RuntimeConfig {
	// 保守默认值：保证系统能跑，但行为偏“安全”
	return &RuntimeConfig{
		OrderCancelTimeoutSec: 300, // 5分钟，按你业务可调
		ProductStreamKey:      "product_events",
		ProductGroup:          "product_group",
	}
}

func (c *RuntimeConfig) WithDefaults() *RuntimeConfig {
	d := DefaultRuntimeConfig()

	if c == nil {
		return d
	}

	// 复制一份，避免修改原对象（并发安全/可预期）
	out := *c

	if out.OrderCancelTimeoutSec <= 0 {
		out.OrderCancelTimeoutSec = d.OrderCancelTimeoutSec
	}
	if out.ProductStreamKey == "" {
		out.ProductStreamKey = d.ProductStreamKey
	}
	if out.ProductGroup == "" {
		out.ProductGroup = d.ProductGroup
	}
	return &out
}

func SetRuntimeConfig(cfg *RuntimeConfig) {
	runtimeCfg.Store(cfg)
}

func GetRuntimeConfig() *RuntimeConfig {
	cfg := runtimeCfg.Load()
	if cfg == nil {
		return DefaultRuntimeConfig()
	}
	return cfg.(*RuntimeConfig).WithDefaults()
}
