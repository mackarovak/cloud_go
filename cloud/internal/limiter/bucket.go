package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity       int
	tokens         int
	refillRate     int
	lastRefill     time.Time
	refillInterval time.Duration
	mu             sync.Mutex
}

func NewTokenBucket(capacity, refillRate int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillRate:     refillRate,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	refillCount := int(elapsed / tb.refillInterval) * tb.refillRate
	
	if refillCount > 0 {
		tb.tokens = min(tb.tokens+refillCount, tb.capacity)
		tb.lastRefill = now
	}
}

func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.refill()
	
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}