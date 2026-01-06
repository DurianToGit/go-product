package redislock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Locker {
	return &Locker{rdb: rdb}
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (l *Locker) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	token, err := randomToken()
	if err != nil {
		return "", false, err
	}

	ok, err := l.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	if !ok {
		return "", false, nil
	}
	return token, true, nil
}

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

func (l *Locker) Unlock(ctx context.Context, key, token string) (bool, error) {
	res, err := unlockScript.Run(ctx, l.rdb, []string{key}, token).Int64()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
