package redis

import (
	"context"
	"product-service/pkg/breaker"
	"time"
)

var Breaker *breaker.CircuitBreaker

func SetBreaker(cb *breaker.CircuitBreaker) {
	Breaker = cb
}

// Do : 统一包装 Redis 调用（带熔断）
func Do(ctx context.Context, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	if Breaker == nil {
		return fn()
	}
	return Breaker.Execute(fn)
}
