package delayqueue

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type DelayQueue struct {
	rdb *redis.Client
	key string
}

func NewDelayQueue(rdb *redis.Client, key string) *DelayQueue {
	return &DelayQueue{
		rdb: rdb,
		key: key,
	}
}

// 入队（延迟任务）
func (q *DelayQueue) Add(ctx context.Context, task interface{}, executeAt int64) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return q.rdb.ZAdd(ctx, q.key, redis.Z{
		Score:  float64(executeAt),
		Member: data,
	}).Err()
}

// 取出到期任务
func (q *DelayQueue) PopReady(ctx context.Context, now int64, limit int64) ([]string, error) {
	return q.rdb.ZRangeByScore(ctx, q.key, &redis.ZRangeBy{
		Min:   "0",
		Max:   string(rune(now)),
		Count: limit,
	}).Result()
}

// 删除任务
func (q *DelayQueue) Remove(ctx context.Context, members []string) error {
	interfaceMembers := make([]interface{}, len(members))
	for i, m := range members {
		interfaceMembers[i] = m
	}
	return q.rdb.ZRem(ctx, q.key, interfaceMembers...).Err()
}
