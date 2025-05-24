package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM)
	defer stop()

	// prepare singleton / global logging service
	log.Logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Ctx(ctx).Logger()

	// cleanup tracing on shutdown
	defer func() {
		if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("error shutting down tracer provider: %v", err)
			}
		}
	}()

	// shared middlewares
	logging := partial(hlog.NewHandler(log.Logger), hlog.URLHandler("path"), hlog.MethodHandler("method"), hlog.RemoteIPHandler("remote"), AccessLog)
	tracing := partial(RequestCounter, RequestLatency)

	// set up an HTTP server
	mux := http.NewServeMux()
	mux.Handle("/joke", chain(FetchJoke(), logging, tracing))
	mux.Handle("/random", chain(RandomNumber(), logging, tracing))
	mux.Handle("/health", HealthEndpoint())
	mux.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = fmt.Sprintf(":%s", port)
	}

	log.Info().Msgf("starting server on %s", addr)
	server := &http.Server{Addr: addr, Handler: mux, BaseContext: func(net.Listener) context.Context { return ctx }}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msgf("server failed to start: %v", err)
		}
	}()

	<-ctx.Done() // wait for signal to terminate
}
