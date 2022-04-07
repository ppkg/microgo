package sys

import (
	"context"

	"github.com/maybgit/glog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracerProvider() {
	if _opt.JaegerEndpoint == "" {
		return
	}

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

func UseTrace(f func()) {
	if _opt != nil && _opt.JaegerEndpoint != "" {
		f()
	}
}

func TraceIsEnable() bool {
	return _opt != nil && _opt.JaegerEndpoint != ""
}
