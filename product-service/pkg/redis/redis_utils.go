package redis

import (
	"context"
	"time"
)

func GetBytes(ctx context.Context, key string) ([]byte, error) {
	rctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	var (
		b   []byte
		err error
	)
	err = Do(rctx, func() error {
		b, err = Client.Get(rctx, key).Bytes()
		return err
	})
	return b, err
}

func SetBytes(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	rctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err := Do(rctx, func() error {
		return Client.Set(rctx, key, val, expiration).Err()
	})
	return err
}

func Delete(ctx context.Context, key string) error {
	rctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err := Do(rctx, func() error {
		return Client.Del(rctx, key).Err()
	})
	return err
}
