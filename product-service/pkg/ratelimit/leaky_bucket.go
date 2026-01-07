package ratelimit

import (
	"sync"
	"time"
)

// LeakyBucket 漏桶结构
type LeakyBucket struct {
	capacity     float64    // 桶容量
	rate         float64    // 漏水速率 (单位/秒)
	water        float64    // 当前水量
	lastLeakTime time.Time  // 上次漏水时间
	mutex        sync.Mutex // 并发控制
}

// NewLeakyBucket 创建漏桶
func NewLeakyBucket(capacity, rate float64) *LeakyBucket {
	return &LeakyBucket{
		capacity:     capacity,
		rate:         rate,
		water:        0,
		lastLeakTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (lb *LeakyBucket) Allow() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 模拟漏水
	now := time.Now()
	elapsed := now.Sub(lb.lastLeakTime).Seconds()
	leakedWater := elapsed * lb.rate

	// 更新水量和时间
	lb.water = max(0, lb.water-leakedWater)
	lb.lastLeakTime = now

	// 检查是否能添加新请求
	if lb.water+1 <= lb.capacity {
		lb.water++
		return true
	}
	return false
}
