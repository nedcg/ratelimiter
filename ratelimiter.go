package ratelimiter

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow() bool
}

type ClockFunc func() time.Time

type ConfigParams struct {
	Tokens     int
	RefillRate int
	Clock      ClockFunc
}

type Config[T RateLimiter] struct {
	*ConfigParams
	TokenGeneratorFunc func(ConfigParams) T
}

type rateLimiterController[T RateLimiter] struct {
	registry map[string]RateLimiter

	// config
	*Config[T]

	in  chan string
	out chan bool
}

func New[T RateLimiter](context context.Context, config Config[T]) rateLimiterController[T] {
	in, out := make(chan string), make(chan bool)

	tbm := rateLimiterController[T]{
		registry: make(map[string]RateLimiter),
		Config:   &config,
		in:       in,
		out:      out,
	}

	tbm.init(context)

	return tbm
}

func (tbm *rateLimiterController[T]) init(ctx context.Context) {
	go func() {
		for {
			select {
			case key := <-tbm.in:
				tbm.out <- tbm.allow(key)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (tbm *rateLimiterController[T]) allow(k string) bool {
	var bucket RateLimiter

	if bucket = tbm.registry[k]; bucket == nil {
		// create new token bucket
		bucket = tbm.TokenGeneratorFunc(ConfigParams{
			Tokens:     tbm.Tokens,
			RefillRate: tbm.RefillRate,
			Clock:      tbm.Clock,
		})

		// add to registry
		tbm.registry[k] = bucket
	}

	return bucket.Allow()
}

func (tbm *rateLimiterController[T]) Allow(k string) bool {
	tbm.in <- k
	return <-tbm.out
}
