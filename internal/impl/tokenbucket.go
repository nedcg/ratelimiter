package impl

import (
	"github.com/nedcg/ratelimiter/pkg/config"
	"time"
)

type TokenBucket struct {
	tokens     int
	bucketCap  int
	refillRate int
	lastRefill time.Time
	clock      config.ClockFunc
}

func NewTokenBucket(config config.ConfigParams) TokenBucket {
	return TokenBucket{
		tokens:     config.Tokens,
		bucketCap:  config.Tokens,
		refillRate: config.RefillRate,
		lastRefill: config.Clock(),
		clock:      config.Clock,
	}
}

func (tb TokenBucket) Allow() bool {
	// before leaking, try to refill the bucketCap
	tb.refill()

	// Comform if there are tokens
	if tb.tokens > 0 {
		tb.tokens -= 1
		return true
	}

	return false
}

func (tb TokenBucket) refill() {
	now := tb.clock()
	elapsed := tb.clock().Sub(tb.lastRefill)

	if elapsed.Seconds() < float64(tb.refillRate) {
		// not enough time passed to refill
		return
	}

	tb.lastRefill = now

	tb.tokens += int(elapsed.Seconds()) * tb.refillRate
	if tb.tokens > tb.bucketCap {
		// there was an overflow in the bucketCap, reset to bucketCap size
		tb.tokens = tb.bucketCap
	}
}
