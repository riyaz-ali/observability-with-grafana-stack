package main

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"net/http"
	"time"
)

type MiddlewareFn func(http.Handler) http.Handler

func chain(endpoint http.Handler, middlewares ...MiddlewareFn) http.Handler {
	if len(middlewares) == 0 {
		return endpoint
	}

	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

func partial(middlewares ...MiddlewareFn) MiddlewareFn {
	return func(next http.Handler) http.Handler {
		return chain(next, middlewares...)
	}
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zerolog.Ctx(r.Context()).Info().Send()

		next.ServeHTTP(w, r)
	})
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
