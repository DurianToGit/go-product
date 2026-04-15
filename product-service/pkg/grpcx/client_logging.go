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

		err := invoker(ctx, method, req, reply, cc, opts...)

		cost := time.Since(start)

		fields := []zap.Field{
			zap.String("grpc_method", method),
			zap.Duration("latency", cost),
		}

		if err != nil {
			logger.L().Error("grpc client request failed", append(fields, zap.Error(err))...)
			return err
		}

		logger.L().Info("grpc client request success", fields...)
		return nil
	}
}
