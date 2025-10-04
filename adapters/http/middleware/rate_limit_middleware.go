package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting per IP address
type RateLimiter struct {
	buckets     map[string]*TokenBucket
	mutex       sync.RWMutex
	rate        int           // requests per minute
	capacity    int           // bucket capacity
	cleanup     time.Duration // cleanup interval for expired buckets
	lastCleanup time.Time     // last cleanup timestamp
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int       // current token count
	capacity   int       // maximum token capacity
	rate       int       // refill rate (tokens per minute)
	lastRefill time.Time // last refill timestamp
	mutex      sync.Mutex
}

// NewRateLimiter creates a new rate limiter with specified rate (requests per minute)
func NewRateLimiter(requestsPerMinute, bucketCapacity int) *RateLimiter {
	return &RateLimiter{
		buckets:     make(map[string]*TokenBucket),
		rate:        requestsPerMinute,
		capacity:    bucketCapacity,
		cleanup:     10 * time.Minute, // cleanup every 10 minutes
		lastCleanup: time.Now(),
	}
}

// RateLimit returns a Gin middleware that enforces rate limiting per IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address
		clientIP := getClientIP(c)

		// Check if request is allowed
		allowed, remaining, resetTime, err := rl.IsAllowed(clientIP)
		if err != nil {
			// Log error but don't block the request
			c.Header("X-RateLimit-Error", err.Error())
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// Rate limit exceeded
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": fmt.Sprintf("Too many requests. Limit: %d requests per minute", rl.rate),
					"details": fmt.Sprintf("Rate limit reset at %s", resetTime.Format(time.RFC3339)),
				},
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		// Request allowed, continue
		c.Next()
	}
}

// IsAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) IsAllowed(clientIP string) (allowed bool, remaining int, resetTime time.Time, err error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Periodic cleanup of expired buckets
	if time.Since(rl.lastCleanup) > rl.cleanup {
		rl.cleanupExpiredBuckets()
		rl.lastCleanup = time.Now()
	}

	// Get or create bucket for this IP
	bucket := rl.getBucketForIP(clientIP)

	// Try to consume a token
	allowed, remaining = bucket.ConsumeToken()

	// Calculate reset time (next minute boundary)
	now := time.Now()
	resetTime = now.Truncate(time.Minute).Add(time.Minute)

	return allowed, remaining, resetTime, nil
}

// getBucketForIP gets or creates a token bucket for the given IP
func (rl *RateLimiter) getBucketForIP(ip string) *TokenBucket {
	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = NewTokenBucket(rl.capacity, rl.rate)
		rl.buckets[ip] = bucket
	}
	return bucket
}

// cleanupExpiredBuckets removes buckets that haven't been used recently
func (rl *RateLimiter) cleanupExpiredBuckets() {
	now := time.Now()
	expireThreshold := 30 * time.Minute // Remove buckets unused for 30 minutes

	for ip, bucket := range rl.buckets {
		bucket.mutex.Lock()
		if now.Sub(bucket.lastRefill) > expireThreshold {
			delete(rl.buckets, ip)
		}
		bucket.mutex.Unlock()
	}
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// ConsumeToken attempts to consume a token from the bucket
func (tb *TokenBucket) ConsumeToken() (allowed bool, remaining int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on time elapsed
	tb.refillTokens()

	if tb.tokens > 0 {
		tb.tokens--
		return true, tb.tokens
	}

	return false, 0
}

// refillTokens adds tokens to the bucket based on the elapsed time
func (tb *TokenBucket) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Calculate tokens to add (rate is per minute)
	tokensToAdd := int(elapsed.Minutes() * float64(tb.rate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for load balancers/proxies)
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the list
		if idx := len(forwarded); idx > 0 {
			for i, char := range forwarded {
				if char == ',' || char == ' ' {
					return forwarded[:i]
				}
			}
			return forwarded
		}
	}

	// Check X-Real-IP header (for nginx)
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// GetRateLimitInfo returns current rate limit information for an IP
func (rl *RateLimiter) GetRateLimitInfo(clientIP string) (remaining int, resetTime time.Time) {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	bucket, exists := rl.buckets[clientIP]
	if !exists {
		// No bucket means no requests made yet, so full capacity available
		return rl.capacity, time.Now().Truncate(time.Minute).Add(time.Minute)
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	bucket.refillTokens()
	resetTime = time.Now().Truncate(time.Minute).Add(time.Minute)

	return bucket.tokens, resetTime
}

// Reset clears all rate limit buckets (useful for testing)
func (rl *RateLimiter) Reset() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.buckets = make(map[string]*TokenBucket)
}

// DefaultRateLimiter creates a rate limiter with default settings (5 requests per minute)
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(5, 5) // 5 requests per minute, capacity 5
}

// CustomRateLimit creates a rate limiting middleware with custom parameters
func CustomRateLimit(requestsPerMinute, bucketCapacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, bucketCapacity)
	return limiter.RateLimit()
}
