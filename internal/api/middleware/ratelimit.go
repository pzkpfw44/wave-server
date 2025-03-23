package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/api/response"
)

// Simple in-memory rate limiter
// For production, use a distributed implementation with Redis

// RateLimiter handles rate limiting
type RateLimiter struct {
	logger       *zap.Logger
	requests     map[string][]time.Time
	mutex        sync.RWMutex
	limit        int           // Maximum requests
	window       time.Duration // Time window
	cleanupEvery time.Duration // How often to clean up old records
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration, logger *zap.Logger) *RateLimiter {
	limiter := &RateLimiter{
		logger:       logger.With(zap.String("middleware", "rate_limiter")),
		requests:     make(map[string][]time.Time),
		mutex:        sync.RWMutex{},
		limit:        limit,
		window:       window,
		cleanupEvery: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// cleanup periodically removes old request timestamps
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupEvery)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for ip, times := range rl.requests {
			var newTimes []time.Time
			for _, t := range times {
				if time.Since(t) < rl.window {
					newTimes = append(newTimes, t)
				}
			}
			if len(newTimes) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = newTimes
			}
		}
		rl.mutex.Unlock()
	}
}

// Limit middleware implements rate limiting
func (rl *RateLimiter) Limit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP
			ip := c.RealIP()

			// Check if rate limit exceeded
			now := time.Now()
			rl.mutex.Lock()

			// Initialize if needed
			if _, exists := rl.requests[ip]; !exists {
				rl.requests[ip] = []time.Time{}
			}

			// Remove old timestamps
			var validTimes []time.Time
			for _, t := range rl.requests[ip] {
				if now.Sub(t) < rl.window {
					validTimes = append(validTimes, t)
				}
			}

			// Check rate limit
			if len(validTimes) >= rl.limit {
				rl.mutex.Unlock()
				rl.logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.Int("limit", rl.limit),
					zap.Duration("window", rl.window),
				)

				resp := response.NewErrorResponse(
					"Too many requests. Please try again later.",
					"RATE_LIMIT_EXCEEDED",
				)
				return c.JSON(http.StatusTooManyRequests, resp)
			}

			// Add current timestamp
			rl.requests[ip] = append(validTimes, now)
			rl.mutex.Unlock()

			return next(c)
		}
	}
}

// AuthLimit is a specialized rate limiter for authentication endpoints
func (rl *RateLimiter) AuthLimit() echo.MiddlewareFunc {
	// More restrictive rate limit for auth endpoints
	authRateLimit := NewRateLimiter(20, 5*time.Minute, rl.logger)
	return authRateLimit.Limit()
}
