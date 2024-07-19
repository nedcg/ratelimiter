package pkg

import (
	"context"
	"github.com/nedcg/ratelimiter/internal/impl"
	"github.com/nedcg/ratelimiter/pkg/config"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	now := []time.Time{time.Now().Round(time.Hour)}[0]
	timeTravel := func(duration time.Duration) {
		now = now.Add(duration)
		t.Log("Time travel", now)
	}

	tb := rateLimiterController[impl.TokenBucket]{
		registry: make(map[string]RateLimiter),
		Config: &Config[impl.TokenBucket]{
			ConfigParams: &config.ConfigParams{
				Tokens:     3,
				RefillRate: 1,
				Clock: func() time.Time {
					return now
				},
			},
		},
		in:  nil,
		out: nil,
	}
	tb.init(context.TODO())

	if !tb.Allow("key1") {
		t.Error()
	}
	if !tb.Allow("key1") {
		t.Error()
	}
	if !tb.Allow("key1") {
		t.Error()
	}
	if tb.Allow("key1") {
		t.Error()
	}

	timeTravel(500 * time.Millisecond)

	if tb.Allow("key1") {
		t.Error()
	}

	if !tb.Allow("key2") {
		t.Error()
	}

	timeTravel(500 * time.Millisecond)

	if !tb.Allow("key1") {
		t.Error()
	}
}

func BenchmarkController_TokenBucket(b *testing.B) {
	numGoroutines := []int{1, 5, 100, 1000, 5000}
	for _, p := range numGoroutines {
		b.Run(strconv.Itoa(p), func(b *testing.B) {
			rlCtrl := New(context.Background(), Config[impl.TokenBucket]{
				NewRateLimiterFunc: func(params config.ConfigParams) impl.TokenBucket {
					return impl.NewTokenBucket(params)
				},
			})

			kn := 5_000_000
			keys := make([]string, kn)
			for i := 0; i < kn; i++ {
				keys[i] = "key" + strconv.Itoa(i)
			}

			b.SetParallelism(p)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := "key" + strconv.Itoa(rand.Intn(kn))
					rlCtrl.Allow(key)
				}
			})
			b.ReportAllocs()
		})
	}

}
