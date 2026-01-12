package rediskeys

import "fmt"

// 商品事件流
const ProductStreamKey = "product:stream:event"

func ProductStockLockKey(productId int64) string {
	return fmt.Sprintf("product:lock:stock:%d", productId)
}

func ProductStockKey(productId int64) string {
	return fmt.Sprintf("product:stock:%d", productId)
}

func ProductSearchKey(hash string) string {
	return "product:search:" + hash
}
