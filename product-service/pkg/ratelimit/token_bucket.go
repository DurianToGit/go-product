package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	mu sync.Mutex

	rate  float64 // tokens per second
	burst float64 // capacity

	tokens float64
	last   time.Time
}

func NewTokenBucket(rate float64, burst int) *TokenBucket {
	if rate <= 0 {
		rate = 1
	}
	if burst <= 0 {
		burst = 1
	}
	now := time.Now()
	return &TokenBucket{
		rate:   rate,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   now,
	}
}

func (b *TokenBucket) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// refill
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
		b.last = now
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}
