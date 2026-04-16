package grpcx

import (
	"context"
	"time"

	"product-service/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StreamClientLoggingInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		start := time.Now()

		// 从 context 中取 request_id 并写入 metadata
		rid, ok := ctx.Value(ContextRequestIDKey).(string)
		if ok && rid != "" {
			ctx = WithRequestID(ctx, rid)
		}

		stream, err := streamer(ctx, desc, cc, method, opts...)

		cost := time.Since(start)

		fields := []zap.Field{
			zap.String("grpc_method", method),
			zap.Duration("latency", cost),
			zap.String("request_id", rid),
		}

		if err != nil {
			logger.L().Error("grpc client stream failed", append(fields, zap.Error(err))...)
			return nil, err
		}

		logger.L().Info("grpc client stream started", fields...)
		return stream, nil
	}
}
