package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

var Client *redis.Client

func InitRedis(addr string, password string, db int) *redis.Client {
	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return Client
}
