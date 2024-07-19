package pkg

import (
	"context"
	"github.com/nedcg/ratelimiter/pkg/config"
	"time"
)

type RateLimiter interface {
	Allow() bool
}

type Config[T RateLimiter] struct {
	*config.ConfigParams
	NewRateLimiterFunc func(params config.ConfigParams) T
}

type rateLimiterController[T RateLimiter] struct {
	registry map[string]RateLimiter
	in       chan string
	out      chan bool

	// config
	*Config[T]
}

func New[T RateLimiter](context context.Context, opts Config[T]) rateLimiterController[T] {
	in, out := make(chan string), make(chan bool)

	if opts.ConfigParams == nil {
		opts.ConfigParams = &config.ConfigParams{
			Tokens:     100,
			RefillRate: 10,
		}
	}

	if opts.ConfigParams.Clock == nil {
		opts.ConfigParams.Clock = time.Now
	}

	if opts.ConfigParams.Tokens < 0 || opts.ConfigParams.RefillRate < 0 || opts.NewRateLimiterFunc == nil {
		panic("invalid config params")
	}

	ctrl := rateLimiterController[T]{
		registry: make(map[string]RateLimiter),
		Config:   &opts,
		in:       in,
		out:      out,
	}

	ctrl.init(context)

	return ctrl
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
		bucket = tbm.Config.NewRateLimiterFunc(*tbm.Config.ConfigParams)

		// add to registry
		tbm.registry[k] = bucket
	}

	return bucket.Allow()
}

func (tbm *rateLimiterController[T]) Allow(k string) bool {
	tbm.in <- k
	return <-tbm.out
}
