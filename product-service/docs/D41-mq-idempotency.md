# D41 MQ 基础模型（消费幂等、高可用）

## 1. 当前项目中 Redis Stream 的使用场景
- 产品扣库存事件
- 产品恢复库存事件

## 2. 为什么消息会重复消费
- handler 成功但 ACK 失败
- worker 异常退出
- pending 消息被重新 claim

## 3. 当前幂等方案
- 使用 MySQL 表 `product_event_consumed`
- 使用 `stream + msg_id` 唯一约束去重

## 4. 当前 ACK 时机
- handler 成功后 ACK
- handler 失败不 ACK，消息保留在 pending

## 5. 当前高可用补充
- 使用 `XAUTOCLAIM` 接管长时间未 ACK 的 pending 消息