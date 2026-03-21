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