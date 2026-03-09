## Service 目标（职责）
* 商品 CRUD / 列表 / 搜索

* 商品详情缓存（L1/L2）

* 库存查询（读）

* （后续）库存预热、商品维度运营数据

## Owner 数据表（D37.1 里那套）
* products

* stocks（注意：“库存扣减写”建议归 order-service，product-service 只做展示/聚合读；否则下单链路会强耦合在商品域）
## 对外接口（列 3~5 条）

* GET /v1/products

* GET /v1/products/:id

* GET /v1/products/:id/stock（读）
## 依赖关系（一句话）

---

* D37.4 本阶段迁移product部分，当前阶段暂时依赖 `internal\errno`，`internal\middleware`
* 其余部分在后续继续收拢
* 移除order在 `internal`中的代码
* 移除product在 `internal`中的代码