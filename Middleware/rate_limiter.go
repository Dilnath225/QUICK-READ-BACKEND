package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple in-memory token bucket rate limiter per IP address.
// For production, use Redis-based rate limiting instead.

type visitor struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	rate       float64 // tokens per second
	burst      int     // max tokens
	cleanupInt time.Duration
}

// NewRateLimiter creates a rate limiter. rate = requests per second, burst = max burst size.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		rate:       rate,
		burst:      burst,
		cleanupInt: 3 * time.Minute,
	}

	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(rl.cleanupInt)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastCheck) > rl.cleanupInt {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    float64(rl.burst) - 1,
			lastCheck: now,
		}
		return true
	}