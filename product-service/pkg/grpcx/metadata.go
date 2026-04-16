package grpcx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const MetadataRequestIDKey = "x-request-id"
const ContextRequestIDKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, MetadataRequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(MetadataRequestIDKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func GetRequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextRequestIDKey).(string); ok {
		return v
	}
	return ""
}
