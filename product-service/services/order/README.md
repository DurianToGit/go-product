## Service 目标（职责）
* 订单创建（幂等 IdemKey）

* 库存扣减（写）/ 秒杀扣减（Lua）

* 订单状态流转（创建→支付→取消）

* 异步任务触发（发 Stream/MQ）

* worker/cron 归 order-service（因为消费消息最终影响订单状态/库存一致性）
## Owner 数据表（D37.1 里那套）
* orders

* order_items（如果你拆明细）

* stock_deduct_logs（可选，做幂等/补偿非常加分）
## 对外接口（列 3~5 条）
* POST /v1/orders（幂等）

* GET /v1/orders/:id

* POST /v1/orders/:id/cancel

* POST /v1/orders/:id/pay（D46 才完善）
## 依赖关系（一句话）
order-service 同步调用 product-service：

---

* D37.3 本阶段迁移order写链路