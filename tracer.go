package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"runtime/debug"
)

var meter = otel.Meter("github.com/riyaz-ali/observability-with-grafana")
var tracer = otel.Tracer("github.com/riyaz-ali/observability-with-grafana")

func init() {
	build, _ := debug.ReadBuildInfo()

	svc := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("observability-with-grafana"),
		semconv.ServiceVersion(build.Main.Version),
		attribute.String("environment", "development"),
	)

	exporter := Must(prometheus.New())
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter), sdkmetric.WithResource(svc))
	otel.SetMeterProvider(mp) // set global MeterProvider

	traceExporter := Must(otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("tempo:4317")))
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(svc))
	otel.SetTracerProvider(tp) // set global TracerProvider
}
