package registry

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

type EtcdRegistry struct {
	cli *clientv3.Client
}

func NewEtcdRegistry(endpoints []string) (*EtcdRegistry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdRegistry{
		cli: cli,
	}, nil
}

func (r *EtcdRegistry) Register(ctx context.Context, serviceName string, inst ServiceInstance, ttl int64) error {
	if ttl <= 0 {
		ttl = 10
	}
	// 1) 创建租约
	leaseResp, err := r.cli.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	// 2)写入key 绑定lease
	key := fmt.Sprintf("service/%s/%s", serviceName, inst.ID)
	valByBets, _ := json.Marshal(inst)
	if _, err = r.cli.Put(ctx, key, string(valByBets), clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}
	log.Printf("[registry] Registered %s => %s", key, inst.Addr)

	// 3) keepalive (后台续租）
	kaCh, err := r.cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				// ctx cancel 时 keepalive 会自动停止
				log.Printf("[registry] keepalive canceled: %s/%s", serviceName, inst.ID)
				return
			case ka, ok := <-kaCh:
				if !ok {
					log.Printf("[registry] keepalive channel closed: %s/%s", serviceName, inst.ID)
					return
				}
				_ = ka // 这里只是保持通道消费，必要时可以打 debug
			}
		}
	}()

	return nil
}

func (r *EtcdRegistry) Close() error {
	return r.cli.Close()
}
