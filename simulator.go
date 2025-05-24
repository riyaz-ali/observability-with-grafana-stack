package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"math/rand"
	"time"
)

func init() { rand.Seed(time.Now().UnixNano()) }

// SimulateLatency simulates artificial latency between fn call invocation
func SimulateLatency[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	zerolog.Ctx(ctx).Debug().Ctx(ctx).Msg("simulating latency")
	time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)

	return fn(ctx)
}

// SimulateError returns random error
func SimulateError[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	if rand.Float64() < 0.25 {
		var zero T
		return zero, fmt.Errorf("simulated error")
	}

	return fn(ctx)
}
