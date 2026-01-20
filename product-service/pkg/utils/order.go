package utils

import (
	"fmt"
	"sync/atomic"
	"time"
)

// DistributedOrderGenerator 分布式订单号生成器
type DistributedOrderGenerator struct {
	prefix   string
	sequence int64 // 序列号，防止同一秒内重复
}

// NewDistributedOrderGenerator 创建分布式订单号生成器
func NewDistributedOrderGenerator(prefix string) *DistributedOrderGenerator {
	return &DistributedOrderGenerator{
		prefix: prefix,
	}
}

// GenerateOrderID 生成分布式安全的订单号
func (dog *DistributedOrderGenerator) GenerateOrderID() string {
	now := time.Now()

	// 格式：YYYYMMDDHHMMSS + 4位序列号
	timestamp := now.Format("20060102150405")

	// 原子操作增加序列号，避免并发冲突
	seq := atomic.AddInt64(&dog.sequence, 1) % 10000

	orderID := fmt.Sprintf("%s%s%04d", dog.prefix, timestamp, seq)
	return orderID
}

// GenerateOrderIDWithMicrosecond 使用微秒级别时间精度
func (dog *DistributedOrderGenerator) GenerateOrderIDWithMicrosecond() string {
	now := time.Now()
	timestamp := now.UnixNano() / 1000000 // 微秒时间戳

	seq := atomic.AddInt64(&dog.sequence, 1) % 1000

	orderID := fmt.Sprintf("%s%d%03d", dog.prefix, timestamp, seq)
	return orderID
}
