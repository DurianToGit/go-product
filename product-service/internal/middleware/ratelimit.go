package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"product-service/internal/errno"
	"product-service/pkg/ratelimit"
	"product-service/pkg/response"
	"sync"
	"time"
)

type limiterEntry struct {
	limiter  *ratelimit.TokenBucket
	lastSeek time.Time
}

type LimiterPool struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	ttl      time.Duration
}

func NewLimiterPool() *LimiterPool {
	pool := &LimiterPool{
		limiters: make(map[string]*limiterEntry),
		ttl:      time.Minute * 10,
	}
	go pool.startCleanup()
	return pool
}

func (l *LimiterPool) Get(key string, rate float64, burst int) *ratelimit.TokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	limiterE, ok := l.limiters[key]
	if !ok {
		limiter := ratelimit.NewTokenBucket(rate, burst)
		l.limiters[key] = &limiterEntry{
			limiter:  limiter,
			lastSeek: time.Now(),
		}
		return limiter
	}
	limiterE.lastSeek = time.Now()
	return limiterE.limiter
}

func (l *LimiterPool) startCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.cleanupExpired()
	}
}

func (l *LimiterPool) cleanupExpired() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.limiters {
		if time.Now().Sub(v.lastSeek) > l.ttl {
			delete(l.limiters, k)
		}
	}
}

func RateLimit(name string, rate float64, burst int, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	pool := NewLimiterPool()
	return func(c *gin.Context) {
		key := name + ":" + keyFunc(c)
		limiter := pool.Get(key, rate, burst)
		if !limiter.Allow(time.Now()) {
			response.ErrorWithErrno(c, errno.ErrTooManyRequests)
			log.Printf("too many requests: %s", key)
			c.Abort()
			return
		}
		c.Next()
	}

}
