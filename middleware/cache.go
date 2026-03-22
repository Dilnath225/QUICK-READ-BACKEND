package middleware

import (
	"QUICK-READ-BACKEND/config"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheResponse is a Gin middleware that caches GET responses in Redis.
// ttl controls how long the cached response lives.
func CacheResponse(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip if Redis is not available
		if config.RedisClient == nil {
			c.Next()
			return
		}

		// Generate cache key from the full URL
		key := "cache:" + hashKey(c.Request.RequestURI)

		// Try to get from cache
		cached, err := config.CacheGet(key)
		if err == nil && cached != "" {
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
			c.Abort()
			return
		}

		// Use a response writer wrapper to capture the response
		w := &responseCapture{ResponseWriter: c.Writer, body: []byte{}}
		c.Writer = w

		c.Next()

		// Only cache successful responses
		if c.Writer.Status() == http.StatusOK && len(w.body) > 0 {
			_ = config.CacheSet(key, string(w.body), ttl)
		}

		c.Header("X-Cache", "MISS")
	}
}

func hashKey(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

// responseCapture wraps gin.ResponseWriter to capture the response body
type responseCapture struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseCapture) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}
