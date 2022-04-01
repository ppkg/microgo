package sys

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"github.com/maybgit/glog"
)

func InitTracerProvider() {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(_opt.JaegerEndpoint)))
	if err != nil {
		glog.Error(err)
		return
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(_opt.Name),
			// attribute.String("ID", uuid.New().String()),
		)),
	)
	otel.SetTracerProvider(tp)
}

func StartTrace(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	var span trace.Span
	ctx, span = otel.Tracer(_opt.Name).Start(ctx, spanName, opts...)
	return ctx, span
}
