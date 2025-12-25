# Product Service


Product Service 是一个使用 Go、Gin、GORM、MySQL 构建的微服务，提供产品的基础增删改查能力，并可作为多微服务系统的基础模板。

## 目录结构
 ```
 product-service/
├── cmd/
│   └── api/
│       └── main.go          // 程序入口
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