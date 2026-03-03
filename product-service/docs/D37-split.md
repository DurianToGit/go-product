## A. user-service（用户域）

###职责

* 注册/登录/JWT

* 用户资料、用户状态

* （后续）用户地址、权限

### 拥有数据

* users / user_profiles（你的项目里具体表名按现有来）

### 对外接口（HTTP 先，D49+ 再升级 gRPC）

* POST /v1/auth/login

* POST /v1/auth/register

* GET /v1/users/me

* GET /v1/users/:id（可选）

##B. product-service（商品域）

### 职责

* 商品 CRUD / 列表 / 搜索

* 商品详情缓存（L1/L2）

* 库存查询（读）

* （后续）库存预热、商品维度运营数据

### 拥有数据

* products

* stocks（注意：“库存扣减写”建议归 order-service，product-service 只做展示/聚合读；否则下单链路会强耦合在商品域）

### 对外接口

* GET /v1/products

* GET /v1/products/:id

* GET /v1/products/:id/stock（读）

## C. order-service（订单域，写链路核心）

### 职责

* 订单创建（幂等 IdemKey）

* 库存扣减（写）/ 秒杀扣减（Lua）

* 订单状态流转（创建→支付→取消）

* 异步任务触发（发 Stream/MQ）

* worker/cron 归 order-service（因为消费消息最终影响订单状态/库存一致性）

### 拥有数据

* orders

* order_items（如果你拆明细）

* stock_deduct_logs（可选，做幂等/补偿非常加分）

### 对外接口

* POST /v1/orders（幂等）

* GET /v1/orders/:id

* POST /v1/orders/:id/cancel

* POST /v1/orders/:id/pay（D46 才完善）

## 服务间依赖关系
   ### 订单创建（同步 + 异步）

* order-service 同步调用 product-service：
* GetProduct(productID) / GetStock(productID)（读、校验）

* order-service 异步发事件：
* OrderCreated → worker 做通知/埋点/延迟取消等（你现在用 Redis Stream，后面 D41+ 可换 Kafka）

> 关键点：库存扣减属于 order-service 的写链路，这样“下单+扣减+幂等”能在一个服务内闭环。