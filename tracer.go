package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"net/http"
	"time"
)

var meter = otel.Meter("github.com/riyaz-ali/observability-with-grafana")
var tracer = otel.Tracer("github.com/riyaz-ali/observability-with-grafana")

func init() {
	svc := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("observability-with-grafana"),
		semconv.ServiceVersion("1.0.0"),
		attribute.String("environment", "development"),
	)

	exporter := Must(prometheus.New())
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter), sdkmetric.WithResource(svc))
	otel.SetMeterProvider(mp) // set global MeterProvider

	traceExporter := Must(otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("tempo:4317")))
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(svc))
	otel.SetTracerProvider(tp) // set global TracerProvider
}

func RequestCounter(next http.Handler) http.Handler {
	counter := Must(meter.Int64Counter("http_requests_total", metric.WithDescription("Total number of HTTP requests.")))
	active := Must(meter.Int64UpDownCounter("http_active_requests", metric.WithDescription("Number of active HTTP requests.")))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attr := metric.WithAttributes(attribute.String("method", r.Method), attribute.String("path", r.URL.Path))
		counter.Add(r.Context(), 1, attr)

		active.Add(r.Context(), 1)
		defer active.Add(r.Context(), -1)

		next.ServeHTTP(w, r)
	})
}

func RequestLatency(next http.Handler) http.Handler {
	latency := Must(meter.Float64Histogram("http_request_latency_seconds",
		metric.WithDescription("HTTP request latency in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
	))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			latency.Record(r.Context(), duration, metric.WithAttributes(attribute.String("path", r.URL.Path)))
		}()

		next.ServeHTTP(w, r)
	})
}
