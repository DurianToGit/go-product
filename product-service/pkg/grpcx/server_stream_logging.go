package grpcx

import (
	"time"

	"product-service/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StreamServerLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		err := handler(srv, ss)

		cost := time.Since(start)

		rid := GetRequestIDFromContext(ss.Context())

		fields := []zap.Field{
			zap.String("grpc_method", info.FullMethod),
			zap.Duration("latency", cost),
			zap.String("request_id", rid),
		}

		if err != nil {
			logger.L().Error("grpc server stream failed", append(fields, zap.Error(err))...)
			return err
		}

		logger.L().Info("grpc server stream success", fields...)
		return nil
	}
}
