package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PostRateLimiter provides enhanced rate limiting specifically for post operations
type PostRateLimiter struct {
	generalLimiter    *RateLimiter
	postCreateLimiter *RateLimiter
	uploadLimiter     *RateLimiter
	bulkLimiter       *RateLimiter
	searchLimiter     *RateLimiter
	operationCounts   map[string]*OperationCounter
	mutex             sync.RWMutex
}

// OperationCounter tracks operation counts per IP for different time windows
type OperationCounter struct {
	hourlyPosts int
	dailyPosts  int
	weeklyPosts int
	lastReset   time.Time
	hourlyReset time.Time
	dailyReset  time.Time
	weeklyReset time.Time
	mutex       sync.Mutex
}

// NewPostRateLimiter creates a new post-specific rate limiter
func NewPostRateLimiter() *PostRateLimiter {
	return &PostRateLimiter{
		generalLimiter:    NewRateLimiter(60, 60), // 60 requests per minute
		postCreateLimiter: NewRateLimiter(10, 15), // 10 posts per minute, burst of 15
		uploadLimiter:     NewRateLimiter(5, 10),  // 5 uploads per minute, burst of 10
		bulkLimiter:       NewRateLimiter(2, 5),   // 2 bulk operations per minute, burst of 5
		searchLimiter:     NewRateLimiter(30, 50), // 30 searches per minute, burst of 50
		operationCounts:   make(map[string]*OperationCounter),
	}
}

// GeneralPostRateLimit applies general rate limiting to all post endpoints
func (prl *PostRateLimiter) GeneralPostRateLimit() gin.HandlerFunc {
	return prl.generalLimiter.RateLimit()
}

// PostCreationRateLimit applies stricter limits to post creation endpoints
func (prl *PostRateLimiter) PostCreationRateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)

		// Check rate limit
		allowed, remaining, resetTime, err := prl.postCreateLimiter.IsAllowed(clientIP)
		if err != nil {
			c.Header("X-RateLimit-Error", err.Error())
		}

		// Set headers
		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		c.Header("X-RateLimit-Type", "post-creation")

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Post creation rate limit exceeded",
				"limit":       "10 posts per minute",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		// Check daily/weekly limits for post creation
		if !prl.checkPostCreationLimits(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily or weekly post creation limit exceeded",
				"limits": gin.H{
					"hourly": "50 posts per hour",
					"daily":  "200 posts per day",
					"weekly": "1000 posts per week",
				},
			})
			c.Abort()
			return
		}

		// Increment post creation counter
		prl.incrementPostCount(clientIP)

		c.Next()
	})
}

// FileUploadRateLimit applies rate limiting to file upload endpoints
func (prl *PostRateLimiter) FileUploadRateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)

		allowed, remaining, resetTime, err := prl.uploadLimiter.IsAllowed(clientIP)
		if err != nil {
			c.Header("X-RateLimit-Error", err.Error())
		}

		c.Header("X-RateLimit-Limit", "5")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		c.Header("X-RateLimit-Type", "file-upload")

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "File upload rate limit exceeded",
				"limit":       "5 uploads per minute",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// BulkOperationRateLimit applies strict limits to bulk operations
func (prl *PostRateLimiter) BulkOperationRateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)

		allowed, remaining, resetTime, err := prl.bulkLimiter.IsAllowed(clientIP)
		if err != nil {
			c.Header("X-RateLimit-Error", err.Error())
		}

		c.Header("X-RateLimit-Limit", "2")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		c.Header("X-RateLimit-Type", "bulk-operation")

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Bulk operation rate limit exceeded",
				"limit":       "2 bulk operations per minute",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// SearchRateLimit applies rate limiting to search endpoints
func (prl *PostRateLimiter) SearchRateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)

		allowed, remaining, resetTime, err := prl.searchLimiter.IsAllowed(clientIP)
		if err != nil {
			c.Header("X-RateLimit-Error", err.Error())
		}

		c.Header("X-RateLimit-Limit", "30")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		c.Header("X-RateLimit-Type", "search")

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Search rate limit exceeded",
				"limit":       "30 searches per minute",
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// DynamicRateLimit applies different limits based on endpoint type
func (prl *PostRateLimiter) DynamicRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Determine rate limit type based on endpoint
		switch {
		case strings.Contains(path, "/posts") && method == "POST":
			prl.PostCreationRateLimit()(c)
		case strings.Contains(path, "/attachments") && method == "POST":
			prl.FileUploadRateLimit()(c)
		case strings.Contains(path, "/bulk"):
			prl.BulkOperationRateLimit()(c)
		case strings.Contains(path, "/search"):
			prl.SearchRateLimit()(c)
		default:
			prl.GeneralPostRateLimit()(c)
		}
	}
}

// checkPostCreationLimits checks hourly, daily, and weekly post creation limits
func (prl *PostRateLimiter) checkPostCreationLimits(clientIP string) bool {
	prl.mutex.Lock()
	defer prl.mutex.Unlock()

	counter, exists := prl.operationCounts[clientIP]
	if !exists {
		counter = &OperationCounter{
			lastReset:   time.Now(),
			hourlyReset: time.Now().Truncate(time.Hour).Add(time.Hour),
			dailyReset:  time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
			weeklyReset: time.Now().Truncate(7 * 24 * time.Hour).Add(7 * 24 * time.Hour),
		}
		prl.operationCounts[clientIP] = counter
	}

	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	now := time.Now()

	// Reset counters if time windows have passed
	if now.After(counter.hourlyReset) {
		counter.hourlyPosts = 0
		counter.hourlyReset = now.Truncate(time.Hour).Add(time.Hour)
	}

	if now.After(counter.dailyReset) {
		counter.dailyPosts = 0
		counter.dailyReset = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	}

	if now.After(counter.weeklyReset) {
		counter.weeklyPosts = 0
		counter.weeklyReset = now.Truncate(7 * 24 * time.Hour).Add(7 * 24 * time.Hour)
	}

	// Check limits
	return counter.hourlyPosts < 50 && counter.dailyPosts < 200 && counter.weeklyPosts < 1000
}

// incrementPostCount increments post creation counters
func (prl *PostRateLimiter) incrementPostCount(clientIP string) {
	prl.mutex.Lock()
	defer prl.mutex.Unlock()

	counter, exists := prl.operationCounts[clientIP]
	if !exists {
		return // Should exist from checkPostCreationLimits call
	}

	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	counter.hourlyPosts++
	counter.dailyPosts++
	counter.weeklyPosts++
}

// AdminBypassRateLimit allows admin users to bypass rate limits
func (prl *PostRateLimiter) AdminBypassRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin (would require JWT parsing)
		// For now, check for admin header or specific API key
		if isAdminRequest(c) {
			c.Header("X-RateLimit-Bypassed", "admin")
			c.Next()
			return
		}

		// Apply normal rate limiting
		prl.DynamicRateLimit()(c)
	}
}

// isAdminRequest checks if the request is from an admin user
func isAdminRequest(c *gin.Context) bool {
	// Check for admin API key
	apiKey := c.GetHeader("X-Admin-Key")
	if apiKey != "" {
		// In a real implementation, validate this against a secure admin key
		return apiKey == "admin-bypass-key" // placeholder
	}

	// Check for admin role in JWT (requires JWT middleware to have run first)
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr == "admin" || roleStr == "super_admin"
		}
	}

	return false
}

// GetRateLimitStatus returns current rate limit status for debugging
func (prl *PostRateLimiter) GetRateLimitStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		// Get status from all limiters
		generalRemaining, generalReset := prl.generalLimiter.GetRateLimitInfo(clientIP)
		postRemaining, postReset := prl.postCreateLimiter.GetRateLimitInfo(clientIP)
		uploadRemaining, uploadReset := prl.uploadLimiter.GetRateLimitInfo(clientIP)
		bulkRemaining, bulkReset := prl.bulkLimiter.GetRateLimitInfo(clientIP)
		searchRemaining, searchReset := prl.searchLimiter.GetRateLimitInfo(clientIP)

		// Get post creation counts
		prl.mutex.RLock()
		counter, exists := prl.operationCounts[clientIP]
		prl.mutex.RUnlock()

		postCounts := gin.H{
			"hourly": 0,
			"daily":  0,
			"weekly": 0,
		}

		if exists {
			counter.mutex.Lock()
			postCounts["hourly"] = counter.hourlyPosts
			postCounts["daily"] = counter.dailyPosts
			postCounts["weekly"] = counter.weeklyPosts
			counter.mutex.Unlock()
		}

		c.JSON(http.StatusOK, gin.H{
			"client_ip": clientIP,
			"rate_limits": gin.H{
				"general": gin.H{
					"remaining": generalRemaining,
					"reset":     generalReset.Unix(),
					"limit":     60,
				},
				"post_creation": gin.H{
					"remaining": postRemaining,
					"reset":     postReset.Unix(),
					"limit":     10,
				},
				"file_upload": gin.H{
					"remaining": uploadRemaining,
					"reset":     uploadReset.Unix(),
					"limit":     5,
				},
				"bulk_operations": gin.H{
					"remaining": bulkRemaining,
					"reset":     bulkReset.Unix(),
					"limit":     2,
				},
				"search": gin.H{
					"remaining": searchRemaining,
					"reset":     searchReset.Unix(),
					"limit":     30,
				},
			},
			"post_creation_counts": postCounts,
			"post_creation_limits": gin.H{
				"hourly": 50,
				"daily":  200,
				"weekly": 1000,
			},
		})
	}
}

// CleanupExpiredCounters removes expired operation counters
func (prl *PostRateLimiter) CleanupExpiredCounters() {
	prl.mutex.Lock()
	defer prl.mutex.Unlock()

	now := time.Now()
	for ip, counter := range prl.operationCounts {
		counter.mutex.Lock()
		// Remove counters that haven't been used for over a week
		if now.Sub(counter.lastReset) > 7*24*time.Hour {
			delete(prl.operationCounts, ip)
		}
		counter.mutex.Unlock()
	}
}

// DefaultPostRateLimiter creates a post rate limiter with default settings
func DefaultPostRateLimiter() *PostRateLimiter {
	return NewPostRateLimiter()
}

// PostRateLimitMiddleware creates middleware for post-specific rate limiting
func PostRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewPostRateLimiter()
	return limiter.DynamicRateLimit()
}
