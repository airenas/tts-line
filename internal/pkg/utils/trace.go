package utils

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(ServiceName).Start(ctx, name, opts...)
}

func InjectTraceToGRPC(ctx context.Context) context.Context {
	md := metadata.New(nil)
	otel.GetTextMapPropagator().Inject(ctx, grpcMetadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

type grpcMetadataCarrier metadata.MD

func (c grpcMetadataCarrier) Get(key string) string {
	values := metadata.MD(c).Get(strings.ToLower(key))
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c grpcMetadataCarrier) Set(key string, value string) {
	lowerKey := strings.ToLower(key) // gRPC metadata keys are required to be lowercase
	metadata.MD(c).Set(lowerKey, value)
}

func (c grpcMetadataCarrier) Keys() []string {
	var keys []string
	for k := range metadata.MD(c) {
		keys = append(keys, k)
	}
	return keys
}
