package config

import (
	"context"
	"fmt"
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
		Endpoints: endpoints,
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
	cfg := &RuntimeConfig{}

	val, err := l.get(ctx, "/go-product/dev/order/cancel_timeout_sec")
	if err != nil {
		return err
	}
	cfg.OrderCancelTimeoutSec, _ = strconv.ParseInt(val, 10, 64)

	cfg.ProductStreamKey, err = l.get(ctx, "/go-product/dev/stream/product_stream_key")
	if err != nil {
		return err
	}
	cfg.ProductGroup, err = l.get(ctx, "/go-product/dev/stream/product_group")
	if err != nil {
		return err
	}

	SetRuntimeConfig(cfg)
	log.Printf("[etcd] load config: %+v", cfg)
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
	var val string

	err := l.cb.Execute(func() error {
		resp, err := l.cli.Get(ctx, key)
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			// key 缺失属于配置错误，不算 etcd 故障；你可以选择不计入熔断
			return fmt.Errorf("missing etcd key: %s", key)
		}
		val = string(resp.Kvs[0].Value)
		return nil
	})

	return val, err
}
