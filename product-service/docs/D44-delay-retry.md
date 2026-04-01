## 1.为什么立即重试不够
这能跑，但有两个明显问题：

* 会“打爆”下游
* 坏消息会在短时间内疯狂重试
## 2.Kafka 为什么没有原生延迟队列
Kafka 没有原生的延迟队列，因为 Kafka 的设计初衷是处理实时数据，而不是处理延迟数据。
Kafka 不像某些 MQ 一样直接有“消息延迟 5 秒投递”这种原生能力。
所以在 Kafka 里，常见做法不是“设置 delay=5s”，而是：

用多个 retry topic + 不同 consumer 的等待策略，模拟延迟队列。
## 3.当前 demo 的多级 retry topic 方案
定义 3 个 topic：
```
demo.stock.deduct.retry.1s
demo.stock.deduct.retry.5s
demo.stock.deduct.retry.30s
```


再加一个死信 topic：

> demo.stock.deduct.dlq

主 topic 仍然是：

> demo.stock.deduct.requested
## 4.每一级延迟时间
```
demo.stock.deduct.retry.1s:一级 延迟 1s
demo.stock.deduct.retry.5s:二级 延迟 5s
demo.stock.deduct.retry.30s:三级 延迟 30s
```

## 5.最终怎么进入 DLQ
`demo.stock.deduct.retry.30s` 队列失败后，进入 `demo.stock.deduct.dlq`
