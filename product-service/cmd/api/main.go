package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"product-service/internal/middleware"
	"product-service/internal/router"
)

func main() {
	// 创建Gin引擎实例
	r := gin.New()
	
	// 注册中间件：日志、耗时统计、异常恢复
	r.Use(
		middleware.Logger(),  // 日志中间件：记录请求日志
		middleware.Cost(),    // 耗时中间件：统计请求处理时间
		middleware.Recovery(), // 恢复中间件：捕获panic并恢复
	)
	
	// 注册路由：将所有API路由注册到引擎
	router.Register(r)
	
	// 启动HTTP服务器，监听8082端口
	// 如果启动失败，则记录致命错误并退出程序
	if err := r.Run(":8082"); err != nil {
		log.Fatalln(err)
	}
}
