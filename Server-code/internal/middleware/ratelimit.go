package middleware

import (
	"sync"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	count    int
	lastSeen time.Time
}

var limiter = &rateLimiter{
	visitors: make(map[string]*visitor),
}

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.RateLimit.Enabled {
			c.Next()
			return
		}

		key := c.ClientIP()
		path := c.Request.URL.Path

		maxPerMinute := cfg.RateLimit.APIPerMinute
		if path == "/api/v1/auth/login" {
			maxPerMinute = cfg.RateLimit.LoginPerMinute
		}

		limiter.mu.Lock()
		v, exists := limiter.visitors[key]
		if !exists || time.Since(v.lastSeen) > time.Minute {
			limiter.visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()
		currentCount := v.count
		limiter.mu.Unlock()

		if currentCount > maxPerMinute {
			utils.TooManyRequests(c, "")
			c.Abort()
			return
		}

		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Next()
	}
}
