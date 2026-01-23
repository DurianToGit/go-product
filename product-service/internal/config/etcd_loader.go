package config

import (
	"context"
	"log"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdLoader struct {
	cli *clientv3.Client
}

func NewEtcdLoader(endpoints []string) (*EtcdLoader, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdLoader{cli: cli}, nil
}

func (l *EtcdLoader) LoadOnce(ctx context.Context) error {
	cfg := &RuntimeConfig{}

	get := func(key string) string {
		resp, err := l.cli.Get(ctx, key)
		if err != nil || len(resp.Kvs) == 0 {
			log.Fatalf("missing etcd key: %s", key)
		}
		return string(resp.Kvs[0].Value)
	}

	cfg.OrderCancelTimeoutSec, _ = strconv.ParseInt(
		get("/go-product/dev/order/cancel_timeout_sec"), 10, 64)

	cfg.ProductStreamKey = get("/go-product/dev/stream/product_stream_key")
	cfg.ProductGroup = get("/go-product/dev/stream/product_group")

	SetRuntimeConfig(cfg)
	log.Printf("[etcd] load config: %+v", cfg)
	return nil
}

func (l *EtcdLoader) Watch(ctx context.Context) {
	rch := l.cli.Watch(ctx, "/go-product/dev/", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.Printf("[etcd] config changed: %s = %s", ev.Kv.Key, ev.Kv.Value)
			_ = l.LoadOnce(ctx)
		}
	}
}
