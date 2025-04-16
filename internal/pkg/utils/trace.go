package utils

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(ServiceName).Start(ctx, name, opts...)
}

func InjectTraceToGRPC(ctx context.Context) context.Context {
	md := metadata.New(nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}
