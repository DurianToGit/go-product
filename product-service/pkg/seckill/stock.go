package seckill

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"product-service/pkg/rediskeys"
)

//go:embed stock.lua
var luaScriptContent string
var deductScript = redis.NewScript(luaScriptContent)

func DeductStockLua(ctx context.Context, rdb *redis.Client, productID, count int64) (int64, error) {
	key := rediskeys.ProductStockKey(productID)
	res, err := deductScript.Run(ctx, rdb, []string{key}, count).Int64()
	if err != nil {
		return 0, err
	}
	return res, nil
}
