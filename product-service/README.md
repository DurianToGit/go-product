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