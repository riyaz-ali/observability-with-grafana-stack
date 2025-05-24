package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func init() { rand.Seed(time.Now().UnixNano()) }

// FetchJoke fetches a joke from provider and returns a response
func FetchJoke() http.HandlerFunc {
	// client is an otel instrumented http client
	var client = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		ctx, span := tracer.Start(r.Context(), "endpoints.fetch_jokes")
		defer span.End()

		// simulate latency when calling external service
		joke, err := SimulateLatency(ctx, func(ctx context.Context) (string, error) {
			log.Debug().Msgf("requesting dad joke")

			req := Must(http.NewRequestWithContext(ctx, http.MethodGet, "https://icanhazdadjoke.com/", http.NoBody))
			req.Header.Add("Accept", "text/plain") // request plain text response

			resp, err := client.Do(req)
			if err != nil {
				return "", fmt.Errorf("failed to fetch dad joke: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("dad joke API returned non-200 status: %d", resp.StatusCode)
			}

			var body bytes.Buffer
			if _, err = io.Copy(&body, resp.Body); err != nil {
				return "", fmt.Errorf("failed to read dad joke response body: %w", err)
			}

			return body.String(), nil
		})

		if err != nil {
			log.Error().Msgf("failed to fetch dad joke: %v", err)

			span.RecordError(err)                    // record the error in the trace span
			span.SetStatus(codes.Error, err.Error()) // set span status to error
			http.Error(w, "failed to get a joke", http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, joke)
		}
	}
}

// RandomNumber returns a random number, sometimes.
func RandomNumber() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		ctx, span := tracer.Start(r.Context(), "endpoints.do_math")
		defer span.End()

		n, err := SimulateLatency(ctx, func(ctx context.Context) (int, error) {
			return SimulateError(ctx, func(ctx context.Context) (int, error) {
				_, inner := tracer.Start(ctx, "rand.integer")
				defer inner.End()

				return rand.Intn(100), nil
			})
		})

		if err != nil {
			log.Error().Msgf("failed to do math: %v", err)

			span.RecordError(err)                    // record the error in the trace span
			span.SetStatus(codes.Error, err.Error()) // set span status to error
			http.Error(w, "failed to do math", http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, "random number: %d", n)
		}
	}
}
