package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var fixedWindowLua = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
  redis.call("PEXPIRE", key, ttl)
end

if current > limit then
  return 0
end
return 1
`)

type RedisLimiter struct {
	rdb *redis.Client
}

func NewRedisLimiter(rdb *redis.Client) *RedisLimiter {
	return &RedisLimiter{rdb: rdb}
}

func (l *RedisLimiter) Allow(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {

	res, err := fixedWindowLua.Run(
		ctx,
		l.rdb,
		[]string{key},
		limit,
		window.Milliseconds(),
	).Int()

	if err != nil {
		return false, err
	}
	return res == 1, nil
}
