package redis

import (
	"context"
	"time"
)

func GetBytes(ctx context.Context, key string) ([]byte, error) {
	rctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	return Client.Get(rctx, key).Bytes()
}

func SetBytes(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	rctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return Client.Set(rctx, key, val, expiration).Err()
}

func Delete(ctx context.Context, key string) error {
	rctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return Client.Del(rctx, key).Err()
}
