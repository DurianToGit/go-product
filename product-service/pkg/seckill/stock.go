package seckill

import (
	"context"
	_ "embed" // 加载lua
	"github.com/redis/go-redis/v9"
	redisPkg "product-service/pkg/redis"
	"product-service/pkg/rediskeys"
)

//go:embed stock.lua
var luaScriptContent string
var deductScript = redis.NewScript(luaScriptContent)

func DeductStockLua(ctx context.Context, rdb *redis.Client, productID, count int64) (int64, error) {
	var (
		res int64
		err error
	)
	key := rediskeys.ProductStockKey(productID)
	err = redisPkg.Do(ctx, func() error {
		res, err = deductScript.Run(ctx, rdb, []string{key}, count).Int64()
		return err
	})

	return res, nil
}
