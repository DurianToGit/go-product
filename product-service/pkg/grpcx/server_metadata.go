package grpcx

import (
	"context"

	"google.golang.org/grpc"
)

func UnaryServerMetadataInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		rid := GetRequestID(ctx)
		if rid != "" {
			ctx = context.WithValue(ctx, ContextRequestIDKey, rid)
		}
		return handler(ctx, req)
	}
}
