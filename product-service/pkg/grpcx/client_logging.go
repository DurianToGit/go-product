package grpcx

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"product-service/pkg/logger"
)

func UnaryClientLoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		// 这里先尝试从 context 中取 request_id
		// 你可以先约定一个简单方案，比如从 ctx.Value("request_id") 取
		rid, ok := ctx.Value(ContextRequestIDKey).(string)
		if ok && rid != "" {
			ctx = WithRequestID(ctx, rid)
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		cost := time.Since(start)

		fields := []zap.Field{
			zap.String("grpc_method", method),
			zap.Duration("latency", cost),
			zap.String("request_id", rid),
		}

		if err != nil {
			logger.L().Error("grpc client request failed", append(fields, zap.Error(err))...)
			return err
		}

		logger.L().Info("grpc client request success", fields...)
		return nil
	}
}
