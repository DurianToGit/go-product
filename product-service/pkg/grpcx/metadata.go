package grpcx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const RequestIDKey = "x-request-id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(RequestIDKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
