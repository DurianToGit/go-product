package grpcx

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"product-service/pkg/logger"
)

func UnaryServerLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		cost := time.Since(start)

		fields := []zap.Field{
			zap.String("grpc_method", info.FullMethod),
			zap.Duration("latency", cost),
		}

		if err != nil {
			logger.L().Error("grpc server request failed", append(fields, zap.Error(err))...)
			return resp, err
		}

		logger.L().Info("grpc server request success", fields...)
		return resp, nil
	}
}
