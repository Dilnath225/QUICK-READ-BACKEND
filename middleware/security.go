package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common security headers to all responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// MaxBodySize limits the request body size (in bytes).
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = newLimitedReader(c.Request.Body, maxBytes)
		c.Next()
	}
}

// limitedReader wraps an io.ReadCloser with a size limit
type limitedReader struct {
	rc interface {
		Read([]byte) (int, error)
		Close() error
	}
	remaining int64
}

func newLimitedReader(rc interface {
	Read([]byte) (int, error)
	Close() error
}, limit int64) *limitedReader {
	return &limitedReader{rc: rc, remaining: limit}
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.remaining <= 0 {
		return 0, &maxBytesError{}
	}
	if int64(len(p)) > lr.remaining {
		p = p[:lr.remaining]
	}
	n, err := lr.rc.Read(p)
	lr.remaining -= int64(n)
	return n, err
}

func (lr *limitedReader) Close() error {
	return lr.rc.Close()
}

type maxBytesError struct{}

func (e *maxBytesError) Error() string {
	return "request body too large"
}
