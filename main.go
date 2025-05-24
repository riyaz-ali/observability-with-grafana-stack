package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// HealthEndpoint returns the service's health status
func HealthEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "ready")
	}
}

func main() {
	defer func() { // cleanup on shutdown
		if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("error shutting down tracer provider: %v", err)
			}
		}
	}()

	// set up an HTTP server
	http.Handle("/joke", chain(FetchJoke(), RequestCounter, RequestLatency))
	http.Handle("/health", HealthEndpoint())
	http.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = fmt.Sprintf(":%s", port)
	}

	log.Printf("starting server on %s", addr)
	server := &http.Server{Addr: addr}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
