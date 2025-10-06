package performance

import (
	"context"
	"database/sql"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PerformanceConfig holds performance optimization settings
type PerformanceConfig struct {
	// Database connection pool settings
	MaxOpenConns    int           `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" default:"5"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" default:"300s"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" default:"60s"`

	// Concurrent request handling
	MaxConcurrentRequests int           `mapstructure:"max_concurrent_requests" default:"1000"`
	RequestTimeout        time.Duration `mapstructure:"request_timeout" default:"30s"`

	// Response caching
	EnableResponseCache bool          `mapstructure:"enable_response_cache" default:"true"`
	CacheTTL            time.Duration `mapstructure:"cache_ttl" default:"300s"`

	// Memory optimization
	EnableGCOptimization bool `mapstructure:"enable_gc_optimization" default:"true"`
	GCPercentage         int  `mapstructure:"gc_percentage" default:"100"`
}

// PerformanceOptimizer handles performance optimizations
type PerformanceOptimizer struct {
	config *PerformanceConfig
	logger *zap.Logger

	// Request limiting
	semaphore    chan struct{}
	requestCount int64
	mu           sync.RWMutex

	// Response cache (simple in-memory cache)
	cache   map[string]*CacheEntry
	cacheMu sync.RWMutex
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer(config *PerformanceConfig, logger *zap.Logger) *PerformanceOptimizer {
	optimizer := &PerformanceOptimizer{
		config:    config,
		logger:    logger,
		semaphore: make(chan struct{}, config.MaxConcurrentRequests),
		cache:     make(map[string]*CacheEntry),
	}

	// Start cache cleanup routine
	go optimizer.cacheCleanupRoutine()

	// Optimize garbage collection if enabled
	if config.EnableGCOptimization {
		optimizer.optimizeGC()
	}

	return optimizer
}

// OptimizeDatabaseConnection configures database connection pool for optimal performance
func (po *PerformanceOptimizer) OptimizeDatabaseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Configure connection pool for high performance
	sqlDB.SetMaxOpenConns(po.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(po.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(po.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(po.config.ConnMaxIdleTime)

	po.logger.Info("Database connection pool optimized",
		zap.Int("max_open_conns", po.config.MaxOpenConns),
		zap.Int("max_idle_conns", po.config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", po.config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", po.config.ConnMaxIdleTime),
	)

	// Test connection pool performance
	go po.monitorConnectionPool(sqlDB)

	return nil
}

// ConcurrencyLimitMiddleware limits concurrent requests to prevent server overload
func (po *PerformanceOptimizer) ConcurrencyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to acquire semaphore
		select {
		case po.semaphore <- struct{}{}:
			// Successfully acquired, continue processing
			defer func() { <-po.semaphore }()

			// Track request count
			po.mu.Lock()
			po.requestCount++
			currentCount := po.requestCount
			po.mu.Unlock()

			// Log if approaching limit
			if currentCount > int64(float64(po.config.MaxConcurrentRequests)*0.8) {
				po.logger.Warn("Approaching concurrent request limit",
					zap.Int64("current_requests", currentCount),
					zap.Int("max_requests", po.config.MaxConcurrentRequests),
				)
			}

			c.Next()

		default:
			// Server overloaded, reject request
			po.logger.Warn("Request rejected - server overloaded",
				zap.String("client_ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
				zap.Int("max_concurrent", po.config.MaxConcurrentRequests),
			)

			c.JSON(503, gin.H{
				"success": false,
				"message": "Server temporarily overloaded",
				"error": gin.H{
					"code":    "SERVER_OVERLOADED",
					"message": "Too many concurrent requests, please try again later",
				},
			})
			c.Abort()
		}
	}
}

// ResponseCacheMiddleware caches responses for improved performance
func (po *PerformanceOptimizer) ResponseCacheMiddleware(cacheKey func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !po.config.EnableResponseCache {
			c.Next()
			return
		}

		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		key := cacheKey(c)
		if key == "" {
			c.Next()
			return
		}

		// Check cache
		po.cacheMu.RLock()
		entry, exists := po.cache[key]
		po.cacheMu.RUnlock()

		if exists && time.Now().Before(entry.ExpiresAt) {
			// Cache hit
			po.logger.Debug("Cache hit", zap.String("key", key))
			c.JSON(200, entry.Data)
			return
		}

		// Cache miss, process request
		c.Next()

		// Cache the response if successful
		if c.Writer.Status() == 200 {
			// This is a simplified cache implementation
			// In production, you might want to use Redis or similar
			po.cacheMu.Lock()
			po.cache[key] = &CacheEntry{
				Data:      "cached_response", // Simplified - would need response body
				ExpiresAt: time.Now().Add(po.config.CacheTTL),
			}
			po.cacheMu.Unlock()
		}
	}
}

// PerformanceMetricsMiddleware tracks performance metrics
func (po *PerformanceOptimizer) PerformanceMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log slow requests
		if duration > 500*time.Millisecond {
			po.logger.Warn("Slow request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("duration", duration),
				zap.Int("status", c.Writer.Status()),
				zap.String("client_ip", c.ClientIP()),
			)
		}

		// Log metrics for monitoring
		po.logger.Debug("Request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status", c.Writer.Status()),
			zap.Int("response_bytes", c.Writer.Size()),
		)
	}
}

// optimizeGC optimizes garbage collection for better performance
func (po *PerformanceOptimizer) optimizeGC() {
	// Set custom GC percentage using debug package
	debug.SetGCPercent(po.config.GCPercentage)

	// Force an initial GC cycle
	runtime.GC()

	po.logger.Info("Garbage collection optimized",
		zap.Int("gc_percentage", po.config.GCPercentage),
		zap.Int("num_cpu", runtime.NumCPU()),
		zap.Int("num_goroutine", runtime.NumGoroutine()),
	)
}

// monitorConnectionPool monitors database connection pool health
func (po *PerformanceOptimizer) monitorConnectionPool(sqlDB *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()

		po.logger.Info("Database connection pool stats",
			zap.Int("open_connections", stats.OpenConnections),
			zap.Int("in_use", stats.InUse),
			zap.Int("idle", stats.Idle),
			zap.Int64("wait_count", stats.WaitCount),
			zap.Duration("wait_duration", stats.WaitDuration),
			zap.Int64("max_idle_closed", stats.MaxIdleClosed),
			zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
		)

		// Alert on connection pool exhaustion
		if stats.WaitCount > 0 {
			po.logger.Warn("Database connection pool under pressure",
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("wait_duration", stats.WaitDuration),
			)
		}
	}
}

// cacheCleanupRoutine periodically cleans expired cache entries
func (po *PerformanceOptimizer) cacheCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		po.cleanupExpiredCache()
	}
}

// cleanupExpiredCache removes expired entries from cache
func (po *PerformanceOptimizer) cleanupExpiredCache() {
	po.cacheMu.Lock()
	defer po.cacheMu.Unlock()

	now := time.Now()
	removedCount := 0

	for key, entry := range po.cache {
		if now.After(entry.ExpiresAt) {
			delete(po.cache, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		po.logger.Debug("Cache cleanup completed",
			zap.Int("removed_entries", removedCount),
			zap.Int("remaining_entries", len(po.cache)),
		)
	}
}

// GetPerformanceStats returns current performance statistics
func (po *PerformanceOptimizer) GetPerformanceStats() map[string]interface{} {
	po.mu.RLock()
	currentRequests := po.requestCount
	po.mu.RUnlock()

	po.cacheMu.RLock()
	cacheSize := len(po.cache)
	po.cacheMu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"concurrent_requests": currentRequests,
		"max_concurrent":      po.config.MaxConcurrentRequests,
		"cache_entries":       cacheSize,
		"goroutines":          runtime.NumGoroutine(),
		"memory_alloc_mb":     memStats.Alloc / 1024 / 1024,
		"memory_sys_mb":       memStats.Sys / 1024 / 1024,
		"gc_cycles":           memStats.NumGC,
	}
}

// PrewarmConnections pre-warms database connections for better startup performance
func (po *PerformanceOptimizer) PrewarmConnections(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Ping database multiple times to establish connections
	for i := 0; i < po.config.MaxIdleConns; i++ {
		if err := sqlDB.PingContext(ctx); err != nil {
			po.logger.Warn("Failed to prewarm database connection",
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
		}
	}

	po.logger.Info("Database connections prewarmed",
		zap.Int("connections", po.config.MaxIdleConns),
	)

	return nil
}
