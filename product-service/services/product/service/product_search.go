package service

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"math/rand"
	"product-service/pkg/cachekey"
	"product-service/pkg/logger"
	"product-service/pkg/redis"
	"product-service/pkg/rediskeys"
	"product-service/services/product/domain"
	"product-service/services/product/dto"
	"time"
)

var productSearchSF singleflight.Group

type productSearchCache struct {
	List  []*domain.Product `json:"list"`
	Total int64             `json:"total"`
}

func (s *ProductService) ProductSearch(ctx context.Context, req *dto.ProductQuery) ([]*domain.Product, int64, error) {
	cacheHash, err := cachekey.Hash(req)
	if err != nil {
		return s.repo.List(ctx, req)
	}
	key := rediskeys.ProductSearchKey(cacheHash)
	if b, err := redis.GetBytes(ctx, key); err == nil {
		if string(b) == domain.DataCacheNil {
			return []*domain.Product{}, 0, nil
		}
		var cache productSearchCache
		if jerr := json.Unmarshal(b, &cache); jerr == nil {
			return cache.List, cache.Total, nil
		}
		// 反序列化失败：删除坏缓存，回源

		derr := redis.Delete(ctx, key)
		if derr != nil {
			logger.L().Error("删除缓存失败", zap.Error(derr))
		}
	}
	// singleflight 合并回源
	v, err, _ := productSearchSF.Do(key, func() (any, error) {
		if b, err := redis.GetBytes(ctx, key); err == nil {
			if string(b) == domain.DataCacheNil {
				return productSearchCache{List: []*domain.Product{}, Total: 0}, nil
			}
			var cache productSearchCache
			if jerr := json.Unmarshal(b, &cache); jerr == nil {
				return cache, nil
			}
			// 反序列化失败：删除坏缓存，回源
			derr := redis.Delete(ctx, key)
			if derr != nil {
				logger.L().Error("删除缓存失败", zap.Error(derr))
			}
		}
		list, total, err := s.repo.List(ctx, req)
		if err != nil {
			return nil, err
		}
		// 4) 写缓存（TTL 抖动防雪崩）
		ttl := 60*time.Second + time.Duration(rand.Intn(30))*time.Second

		if total == 0 {
			// 空结果短 TTL（防穿透）
			_ = redis.SetBytes(ctx, key, domain.DataCacheNil, 10*time.Second)
			return productSearchCache{List: []*domain.Product{}, Total: 0}, nil
		}

		payload := productSearchCache{List: list, Total: total}
		b, _ := json.Marshal(payload)
		_ = redis.SetBytes(ctx, key, b, ttl)

		return payload, nil
	})
	if err != nil {
		return nil, 0, err
	}
	cached := v.(productSearchCache)
	return cached.List, cached.Total, err
}
