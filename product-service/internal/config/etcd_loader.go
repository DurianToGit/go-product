package config

import (
	"context"
	"log"
	"product-service/pkg/breaker"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdLoader struct {
	cli *clientv3.Client
	cb  *breaker.CircuitBreaker
}

func NewEtcdLoader(endpoints []string) (*EtcdLoader, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 800 * time.Millisecond, // ✅ 关键
	})
	if err != nil {
		return nil, err
	}
	return &EtcdLoader{
		cli: cli,
		cb:  breaker.New(5, 10*time.Second), // 连续失败5次，熔断10秒
	}, nil
}

func (l *EtcdLoader) LoadOnce(ctx context.Context) error {
	cfg := GetRuntimeConfig()

	// 获取每个配置项，并处理可能的错误
	val, err := l.get(ctx, "/go-product/dev/order/cancel_timeout_sec")
	if err == nil {
		cfg.OrderCancelTimeoutSec, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Printf("获取配置项 /go-product/dev/order/cancel_timeout_sec 错误: %v", err)
		}
	}

	cfg.ProductStreamKey, err = l.get(ctx, "/go-product/dev/stream/product_stream_key")
	if err != nil {
		log.Printf("获取配置项 /go-product/dev/stream/product_stream_key 失败: %v", err)
		cfg.ProductStreamKey = DefaultRuntimeConfig().ProductStreamKey
	}

	cfg.ProductGroup, err = l.get(ctx, "/go-product/dev/stream/product_group")
	if err != nil {
		log.Printf("获取配置项 /go-product/dev/stream/product_group 失败: %v", err)
		cfg.ProductGroup = DefaultRuntimeConfig().ProductGroup
	}

	log.Println("准备存储配置")
	SetRuntimeConfig(cfg)
	log.Printf("[etcd] 加载配置完成: %+v", cfg)
	return nil
}

func (l *EtcdLoader) Watch(ctx context.Context) {
	prefix := "/go-product/dev/"
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		rch := l.cli.Watch(ctx, prefix, clientv3.WithPrefix())

		// 一旦进入 watch loop，重置 backoff
		backoff = time.Second

		for wresp := range rch {
			if wresp.Err() != nil {
				log.Printf("[etcd] watch error: %v", wresp.Err())
				break // 跳出内层，外层负责退避重连
			}
			if len(wresp.Events) > 0 {
				// 有事件就 reload（这里走熔断保护的 LoadOnce）
				if err := l.LoadOnce(ctx); err != nil {
					if breaker.IsCircuitOpen(err) {
						log.Printf("[etcd] load skipped (circuit open): %v", err)
					} else {
						log.Printf("[etcd] load failed: %v", err)
					}
					// 保留旧配置（fail-open）
				}
			}

		}

		// watch channel 关闭或出错：退避重连
		log.Printf("[etcd] watch reconnect in %v", backoff)

		t := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (l *EtcdLoader) get(ctx context.Context, key string) (string, error) {
	// ✅ 给每次 Get 一个明确超时，避免启动卡死
	rctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	var val string
	err := l.cb.Execute(func() error {
		resp, err := l.cli.Get(rctx, key)
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			// key 缺失属于配置错误，不算 etcd 故障；你可以选择不计入熔断
			log.Printf("etcd key 不存在: %s，将使用默认值", key)
			return nil
		}
		val = string(resp.Kvs[0].Value)
		return nil
	})

	return val, err
}
