package bootstrap

import (
	"context"
	"net/http"
	"product-service/internal/client/orderclient"
	"product-service/internal/client/productclient"
	"product-service/internal/client/userclient"
	"product-service/internal/gateway/ordergateway"
	"product-service/internal/gateway/productgateway"
	"product-service/internal/gateway/usergateway"
	internalMiddleware "product-service/internal/middleware"
	"product-service/pkg/grpcx"
	"product-service/pkg/logger"
	"product-service/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GatewayApp struct {
	httpServer *http.Server

	userConn    *grpc.ClientConn
	productConn *grpc.ClientConn
	orderConn   *grpc.ClientConn

	userClient    userclient.Client
	productClient productclient.Client
	orderClient   orderclient.Client

	userHandler    *usergateway.Handler
	productHandler *productgateway.Handler
	orderHandler   *ordergateway.Handler
}

func InitGatewayApp() (*GatewayApp, error) {
	cfg := BaseInit()
	app := &GatewayApp{}

	// 资源清理函数列表，用于初始化失败时回滚
	var cleanup []func()

	// 注册清理函数，如果后续初始化失败则执行
	defer func() {
		if cleanup != nil {
			// 初始化失败，执行所有清理函数
			for i := len(cleanup) - 1; i >= 0; i-- {
				cleanup[i]()
			}
		}
	}()

	// 初始化 gRPC 连接
	var productGrpcConn *grpc.ClientConn
	var orderGrpcConn *grpc.ClientConn
	var userGrpcConn *grpc.ClientConn
	var grpcErr error

	// product gRPC 连接
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		productGrpcConn, grpcErr = grpc.DialContext(
			ctx,
			cfg.App.Product.GrpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		)
		cancel()
		if grpcErr != nil {
			logger.L().Error("初始化 product grpc client 失败", zap.Error(grpcErr))
			return nil, grpcErr
		}
		cleanup = append(cleanup, func() {
			if err := productGrpcConn.Close(); err != nil {
				logger.L().Error("关闭 product gRPC 连接失败", zap.Error(err))
			}
		})
	}

	// order gRPC 连接
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		orderGrpcConn, grpcErr = grpc.DialContext(
			ctx,
			cfg.App.Order.GrpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		)
		cancel()
		if grpcErr != nil {
			logger.L().Error("初始化 order grpc client 失败", zap.Error(grpcErr))
			return nil, grpcErr
		}
		cleanup = append(cleanup, func() {
			if err := orderGrpcConn.Close(); err != nil {
				logger.L().Error("关闭 order gRPC 连接失败", zap.Error(err))
			}
		})
	}

	// user gRPC 连接
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		userGrpcConn, grpcErr = grpc.DialContext(
			ctx,
			cfg.App.User.GrpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(grpcx.UnaryClientLoggingInterceptor()),
		)
		cancel()
		if grpcErr != nil {
			logger.L().Error("初始化 user grpc client 失败", zap.Error(grpcErr))
			return nil, grpcErr
		}
		cleanup = append(cleanup, func() {
			if err := userGrpcConn.Close(); err != nil {
				logger.L().Error("关闭 user gRPC 连接失败", zap.Error(err))
			}
		})
	}
	productClient := productclient.NewGRPCClient(productGrpcConn)
	orderClient := orderclient.NewGRPCClient(orderGrpcConn)
	userClient := userclient.NewGRPCClient(userGrpcConn)

	// 创建Gin引擎实例
	r := gin.New()
	// 注册中间件：日志、耗时统计、异常恢复
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(),
		middleware.RecoveryZap(),
		middleware.Cost(), // 耗时中间件：统计请求处理时间
	)

	app.userConn = userGrpcConn
	app.productConn = productGrpcConn
	app.orderConn = orderGrpcConn
	app.userClient = userClient
	app.productClient = productClient
	app.orderClient = orderClient
	app.userHandler = usergateway.NewHandler(userClient)
	app.productHandler = productgateway.NewHandler(productClient)
	app.orderHandler = ordergateway.NewHandler(orderClient)

	// 初始化成功，清空 cleanup 列表，防止 defer 执行清理
	cleanup = nil

	// 注册路由：将所有API路由注册到引擎
	app.RegisterRouter(r)

	// HTTP服务器也需要优雅关闭
	app.httpServer = &http.Server{
		Addr:    cfg.App.Addr,
		Handler: r,
	}
	logger.L().Info("server_listening", zap.String("addr", cfg.App.Addr))

	return app, nil
}

func (a *GatewayApp) Serve() error {
	logger.L().Info("starting_server")
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.L().Error("starting_server_error", zap.Error(err))
		return err
	}
	return nil
}

func (a *GatewayApp) Close() error {
	var errs []error

	// 关闭 gRPC 连接
	for _, conn := range []*grpc.ClientConn{a.productConn, a.orderConn, a.userConn} {
		if conn != nil {
			if err := conn.Close(); err != nil {
				logger.L().Error("关闭 gRPC 连接失败", zap.Error(err))
				errs = append(errs, err)
			}
		}
	}

	// 关闭 HTTP 服务器
	if a.httpServer != nil {
		if err := a.httpServer.Close(); err != nil {
			logger.L().Error("closing_server_error", zap.Error(err))
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (a *GatewayApp) RegisterRouter(r *gin.Engine) {
	api := r.Group("/api/v1")

	usergateway.InitPublicRouter(api, a.userHandler)

	biz := api.Group("")
	biz.Use(internalMiddleware.Auth())
	{
		productgateway.InitRouter(biz, a.productHandler)
		ordergateway.InitRouter(biz, a.orderHandler)
		usergateway.InitRouter(biz, a.userHandler)
	}

}
