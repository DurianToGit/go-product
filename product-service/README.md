# Product Service


Product Service 是一个使用 Go、Gin、GORM、MySQL 构建的微服务，提供产品的基础增删改查能力，并可作为多微服务系统的基础模板。


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