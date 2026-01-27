package registry

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdDiscovery struct {
	cli *clientv3.Client
}

func NewEtcdDiscovery(endpoints []string) (*EtcdDiscovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdDiscovery{cli: cli}, nil
}

func (d *EtcdDiscovery) ListInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)

	resp, err := d.cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	out := make([]ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var inst ServiceInstance
		if err := json.Unmarshal(kv.Value, &inst); err != nil {
			// 忽略坏数据或直接返回错误都行；学习阶段建议返回错误
			return nil, err
		}
		out = append(out, inst)
	}
	return out, nil
}

func (d *EtcdDiscovery) Close() error {
	return d.cli.Close()
}
