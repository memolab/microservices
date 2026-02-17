// Package observe define observing initializing and providers
package observe

import (
	"context"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

func InitGlobalTracer(serviceName, jaegerEndpoint string) (func(context.Context) error, error) {
	// Exporter
	traceExporter, err := newExporter(jaegerEndpoint)
	if err != nil {
		return nil, err
	}

	// Trace Provider
	traceProvider, err := newTraceProvider(serviceName, traceExporter)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(traceProvider)

	// Propagator
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	return traceProvider.Shutdown, nil
}

func newExporter(endpoint string) (sdktrace.SpanExporter, error) {
	// return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	return otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(serviceName string, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return traceProvider, nil
}
