package main

import (
	"bytes"
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
)

// HttpClient is an otel instrumented http client
var HttpClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

// fetchDadJoke makes an HTTP call to icanhazdadjoke.com and returns a joke.
func fetchDadJoke(ctx context.Context) (string, error) {
	req := Must(http.NewRequestWithContext(ctx, http.MethodGet, "https://icanhazdadjoke.com/", http.NoBody))
	req.Header.Add("Accept", "text/plain") // request plain text response

	log.Println("DEBUG: fetching dad joke from external API...")
	resp, err := HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch dad joke: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dad joke API returned non-200 status: %d", resp.StatusCode)
	}

	var body bytes.Buffer
	if _, err = io.Copy(&body, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read dad joke response body: %w", err)
	}

	return body.String(), nil
}

// FetchJoke fetches a joke from provider and returns a response
func FetchJoke() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Printf("INFO: received request for %s", r.URL.Path)

		// start a trace span for the request handler
		ctx, span := tracer.Start(ctx, "jokes.fetch", trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.target", r.URL.Path),
		))
		defer span.End()

		// Fetch a dad joke from the external API
		joke, err := fetchDadJoke(ctx)
		if err != nil {
			log.Printf("ERROR: failed to fetch dad joke: %v", err)
			span.RecordError(err)                    // Record the error in the trace span
			span.SetStatus(codes.Error, err.Error()) // Set span status to error
			http.Error(w, "failed to get a joke", http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintf(w, joke)
		log.Printf("INFO: finished processing request for %s", r.URL.Path)
	}
}
