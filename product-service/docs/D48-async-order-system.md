# D48 异步订单系统完成

## 1. 目标
完成订单系统的异步闭环，包括支付成功事件、取消订单事件、outbox 可靠发布、Kafka 消费、消费幂等。

## 2. 系统角色
- OrderService：负责订单状态流转与事务内写 outbox
- OutboxRelay：负责扫描 pending outbox 并发布 Kafka
- OrderConsumer：负责消费 order.paid / order.canceled
- ProductRepository：负责取消订单后的库存恢复

## 3. 支付成功链路
Pay
-> MarkPaid
-> outbox(order.paid)
-> relay
-> Kafka
-> consumer
-> 异步任务

## 4. 取消订单链路
Cancel / CancelExpired
-> MarkCancelled
-> outbox(order.canceled)
-> relay
-> Kafka
-> consumer
-> RestoreStock

## 5. 为什么需要 Outbox
保证“业务状态变更”和“事件发布”之间的一致性，避免订单状态已经更新但消息未发出的情况。

## 6. 为什么需要消费幂等
Kafka 至少一次投递，消息可能重复消费。
如果不做幂等：
- order.canceled 会重复恢复库存
- order.paid 会重复执行支付后任务

当前方案：
- event_consume_log
- 唯一索引 (event_id, consumer_group)
- 首次消费插入成功才执行业务
- 重复消费直接跳过

补充
- event_id 来源于 outbox_event.id。
- OutboxRelay 发布 Kafka 时，将 outbox 主键写入 message key。
- Consumer 通过 msg.Key 获取 event_id，再写入 event_consume_log 做去重。

## 7. 当前系统能力边界
当前已完成：
- order.paid
- order.canceled
- outbox
- relay
- Kafka consumer
- restore stock
- 消费幂等

后续可继续补强：
- 支付后真实异步任务
- 失败重试监控
- 指标与告警