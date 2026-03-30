package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func InitTracer(serviceName string) (*trace.TracerProvider, error) {
	res := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName(serviceName))
	traceExporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res))
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}
