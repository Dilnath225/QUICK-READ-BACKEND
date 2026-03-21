package middleware

import (
	"QUICK-READ-SYSTEM/config"
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