# Product Service


Product Service 是一个使用 Go、Gin、GORM、MySQL 构建的微服务，提供产品的基础增删改查能力，并可作为多微服务系统的基础模板。

## 目录结构
 ```
 product-service/
├── cmd/
│   └── handler/             // HTTP Handler
├── cmd/
│   ├── api/
│   │   └── main.go          // 程序入口
│   └── worker/
│       └── main.go          // 商品队列消费者服务
├── internal/
│   ├── auth/
│   │   └── jwt.go           // JWT 认证
│   ├── config/
│   │   └── config.go        // 配置文件
│   ├── dto/                 // 接口返回对象
│   │   └── user.go           
│   ├── model(do)/           // 数据库对象
│   │   └── user.go              
│   ├── errno/           
│   │   └── errno.go         // 错误码  
│   ├── router/
│   │   └── router.go        // 路由注册
│   ├── handler/
│   │   └── ping.go          // HTTP Handler 
│   ├── middleware/
│   │   └── logger.go        // 中间件（D12 扩展）
│   ├── response/
│   │   └── response.go      // 统一响应结构
│   ├── service/             // 业务逻辑
│   ├── response/
│   │   └── response.go      // 统一响应结构
│   └── validator/
│       └── validator.go     // 参数验证
├── go.mod

 ```


## 功能列表
- 创建商品
- 查询商品详情
- 查询商品列表
- 更新商品
- 删除商品


---


## 技术栈
- Go 1.22+
- Gin
- GORM
- MySQL 8
- Docker & Docker Compose
- Goose (数据库迁移)


---


## 本地开发


### 1. 启动依赖服务（MySQL）
```bash
docker-compose up -d
```
### 商品服务说明
* 启动流程
  * 先启动Api服务，再启动消费者服务，因为订单服务依赖商品服务，且数据迁移migration在Api服务启动过程完成
* 商品中心包含哪些进程
  * api：主进程，提供 HTTP 接口服务
  * worker：Stream 消费者进程，消费库存扣减事件并幂等落库（避免重复扣减）
  
* Redis 在这里承担什么角色
  * 三个角色：
    1. 缓存：商品搜索结果缓存、用户信息缓存等
    2. 秒杀入口库存：Redis Lua 原子扣减作为高并发入口
    3. 消息队列：Redis Stream 投递扣减事件，解耦主流程与落库

* MySQL 在这里承担什么角色
  * 商品与库存的最终一致持久化（权威数据源），并通过幂等表记录消费结果便于审计/排障

* Stream 在这里解决什么问题
  * 将“Redis 秒杀扣减成功”事件异步投递给 worker，worker 负责 MySQL 最终落库；失败进入 pending 可重试，实现最终一致与解耦

* 服务边界：user/product/order 各自负责什么
  * user 服务：
    * 注册/登录/刷新 token 
    * JWT 签发（AuthN） 
    * 鉴权能力输出：user_id（AuthZ 可先不做复杂） 
    * 用户资料（Profile）
  * product 服务：
    * 商品 CRUD 
    * 库存模型（Redis Lua + MySQL 落库） 
    * 商品相关缓存/搜索/限流 
    * 事件消费幂等（product_event_consumed）
  * order 服务：
    * 订单 CRUD
    * 订单状态机（created/paid/canceled）
    * 超时取消（cron/worker）
    * 订单事件（取消触发恢复库存事件等）

* 依赖方向：谁可以调用谁（禁止循环）
  * 商品和订单服务之间，订单服务依赖商品服务，
  * 用户服务提供用户信息，可以供给商品服务和订单服务使用，商品/订单服务获取 user_id 的方式：本地校验 JWT，不在请求链路上同步调用 user-service（避免把 user 变成瓶颈）

* 数据归属：每张表属于哪个服务；跨服务只传 id，不 join
  * users 表：user 服务
  * products 表：product 服务
  * product_event_consumed 表：product 服务
  * orders 表：order 服务
  * orders.user_id 只存 id，不存用户详情；需要详情由聚合层获取或冗余快照。

### 项目架构问题
  * 为什么服务注册必须和 Lease 绑定，而不能只 put 一个 key？
    * 注册行为的“语义保证”，避免服务注册时，服务实例已经不存在了，但是注册中心没有及时删除，只要进程活着 → key 一直存在 ，进程崩溃 / Kill → Lease 过期 → key 自动删除，你不用手写“下线逻辑”
  * user-service 注册逻辑你准备放在哪里？（main / bootstrap / middleware？为什么）
    * 放在 cmd/api/main.go 或 bootstrap 初始化阶段（BaseInit 之后） 
    * 并且用 signal.NotifyContext 在退出时 cancel（让 KeepAlive goroutine 优雅退出）
  * order-service 如果将来要发现 user-service，你准备从哪个模块调用 discovery？
    * 放在 internal/client 或 internal/service 的“依赖方 client”里，例如： internal/client/userclient（封装 discovery + load balance + http/grpc 调用） 
    * handler 只调用：userClient.GetProfile(ctx, userID) 这种语义化方法 
    > 结论：handler 只做协议适配；discovery 属于 client/infra 层。