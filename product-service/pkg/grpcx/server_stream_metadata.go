package grpcx

import (
	"context"

	"google.golang.org/grpc"
)

func StreamServerMetadataInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		rid := GetRequestID(ctx)
		if rid != "" {
			ctx = context.WithValue(ctx, ContextRequestIDKey, rid)
			ss = &wrappedServerStream{
				ServerStream: ss,
				ctx:          ctx,
			}
		}
		return handler(srv, ss)
	}
}

// wrappedServerStream 包装 ServerStream 以支持自定义 context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
