package cache

import (
	"sync"
	"time"
)

type item struct {
	val       []byte
	expiresAt time.Time
}

type LocalCache struct {
	mu   sync.RWMutex
	data map[string]item
	ttl  time.Duration
}

func NewLocalCache(ttl time.Duration) *LocalCache {
	return &LocalCache{
		data: make(map[string]item),
		ttl:  ttl,
	}
}

func (c *LocalCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	it, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expiresAt) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}
	return it.val, true
}

func (c *LocalCache) Set(key string, val []byte) {
	c.mu.Lock()
	c.data[key] = item{val: val, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *LocalCache) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}
