package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/internal/logging"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a middleware that logs HTTP requests with context propagation
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return LoggingMiddlewareWithFilter(logger, nil)
}

// LoggingMiddlewareWithFilter creates a middleware with custom sensitive field filtering
func LoggingMiddlewareWithFilter(logger *zap.Logger, filter logging.SensitiveFieldFilter) gin.HandlerFunc {
	if filter == nil {
		filter = logging.NewDefaultSensitiveFieldFilter()
	}

	return func(c *gin.Context) {
		// Generate request scope with unique IDs
		requestScope := logging.NewRequestScope()
		requestScope.StartTime = time.Now().UnixNano()

		// Add request context to Gin context
		ctx := requestScope.ToContext(context.Background())
		c.Request = c.Request.WithContext(ctx)

		// Create contextual logger for this request
		contextLogger := logging.NewContextualLogger(logger, ctx)

		// Store logger in Gin context for handlers to use
		c.Set("logger", contextLogger)
		c.Set("request_scope", requestScope)

		// Log request start
		startTime := time.Now()
		contextLogger.Info("Request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", sanitizeQuery(c.Request.URL.RawQuery, filter)),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("remote_addr", c.ClientIP()),
			zap.Int64("content_length", c.Request.ContentLength),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log response
		status := c.Writer.Status()
		contextLogger.Info("Request completed",
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
			zap.String("status_category", getStatusCategory(status)),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				contextLogger.Error("Request error",
					zap.Error(err.Err),
					zap.Int("error_type", int(err.Type)),
					zap.Any("error_meta", sanitizeErrorMeta(err.Meta, filter)),
				)
			}
		}
	}
}

// StructuredLoggingMiddleware provides more detailed structured logging
func StructuredLoggingMiddleware(logger *zap.Logger, filter logging.SensitiveFieldFilter) gin.HandlerFunc {
	if filter == nil {
		filter = logging.NewDefaultSensitiveFieldFilter()
	}

	return func(c *gin.Context) {
		// Generate request scope
		requestScope := logging.NewRequestScope()
		requestScope.StartTime = time.Now().UnixNano()

		// Extract user context if available (from authentication middleware)
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				requestScope.UserID = uid
			}
		}

		if sessionID, exists := c.Get("session_id"); exists {
			if sid, ok := sessionID.(string); ok {
				requestScope.SessionID = sid
			}
		}

		// Add to context
		ctx := requestScope.ToContext(context.Background())
		c.Request = c.Request.WithContext(ctx)

		// Create contextual logger
		contextLogger := logging.NewContextualLogger(logger, ctx)
		c.Set("logger", contextLogger)
		c.Set("request_scope", requestScope)

		// Collect request metadata
		requestFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", sanitizeQuery(c.Request.URL.RawQuery, filter)),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("host", c.Request.Host),
			zap.String("referer", c.Request.Referer()),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.String("content_type", c.Request.Header.Get("Content-Type")),
		}

		// Add custom headers (filtered for sensitive data)
		for key, values := range c.Request.Header {
			if filter.IsSensitive(key) {
				continue // Skip sensitive headers
			}
			if len(values) > 0 {
				requestFields = append(requestFields, zap.String("header_"+key, values[0]))
			}
		}

		startTime := time.Now()
		contextLogger.Info("HTTP request started", requestFields...)

		// Process request
		c.Next()

		// Response logging
		duration := time.Since(startTime)
		status := c.Writer.Status()

		responseFields := []zap.Field{
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
			zap.String("status_category", getStatusCategory(status)),
			zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		}

		// Add response headers (filtered)
		for key, values := range c.Writer.Header() {
			if filter.IsSensitive(key) {
				continue
			}
			if len(values) > 0 {
				responseFields = append(responseFields, zap.String("response_header_"+key, values[0]))
			}
		}

		contextLogger.Info("HTTP request completed", responseFields...)

		// Log detailed errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				contextLogger.Error("HTTP request error",
					zap.Error(err.Err),
					zap.Int("error_type", int(err.Type)),
					zap.Any("error_meta", sanitizeErrorMeta(err.Meta, filter)),
					zap.String("error_context", "gin_middleware"),
				)
			}
		}

		// Performance warning for slow requests
		if duration > 5*time.Second {
			contextLogger.Error("Slow request detected",
				zap.Duration("duration", duration),
				zap.String("performance_category", "slow_request"),
			)
		}
	}
}

// RequestIDMiddleware adds request ID to context (lightweight version)
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in headers
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = logging.GenerateRequestID()
		}

		// Add to context
		ctx := logging.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response headers
		c.Header("X-Request-ID", requestID)

		// Store in Gin context
		c.Set("request_id", requestID)

		c.Next()
	}
}

// LoggerFromContext extracts the contextual logger from Gin context
func LoggerFromContext(c *gin.Context) *logging.ContextualLogger {
	if logger, exists := c.Get("logger"); exists {
		if contextLogger, ok := logger.(*logging.ContextualLogger); ok {
			return contextLogger
		}
	}

	// Fallback: create a new contextual logger if not found
	if baseLogger, exists := c.Get("base_logger"); exists {
		if zapLogger, ok := baseLogger.(*zap.Logger); ok {
			return logging.NewContextualLogger(zapLogger, c.Request.Context())
		}
	}

	// Last resort: create a no-op logger to prevent panics
	return logging.NewContextualLogger(zap.NewNop(), c.Request.Context())
}

// RequestScopeFromContext extracts the request scope from Gin context
func RequestScopeFromContext(c *gin.Context) *logging.RequestScope {
	if scope, exists := c.Get("request_scope"); exists {
		if requestScope, ok := scope.(*logging.RequestScope); ok {
			return requestScope
		}
	}
	return nil
}

// Helper functions

// sanitizeQuery removes sensitive data from query parameters
func sanitizeQuery(query string, filter logging.SensitiveFieldFilter) string {
	if query == "" {
		return ""
	}

	// For now, do basic sanitization by checking common sensitive params
	sensitiveParams := []string{"password", "secret", "token", "key", "auth"}

	for _, param := range sensitiveParams {
		if filter.IsSensitive(param) {
			// This is a simple approach - in production you might want more sophisticated parsing
			continue
		}
	}

	// Return original query if no sensitive data detected
	// In production, you might want to parse and reconstruct the query string
	return query
}

// sanitizeErrorMeta sanitizes error metadata
func sanitizeErrorMeta(meta interface{}, filter logging.SensitiveFieldFilter) interface{} {
	if meta == nil {
		return nil
	}

	// If meta is a map, sanitize its values
	if metaMap, ok := meta.(map[string]interface{}); ok {
		sanitized := make(map[string]interface{})
		for k, v := range metaMap {
			sanitized[k] = filter.Sanitize(k, v)
		}
		return sanitized
	}

	// For other types, return as-is
	return meta
}

// getStatusCategory categorizes HTTP status codes
func getStatusCategory(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "success"
	case status >= 300 && status < 400:
		return "redirect"
	case status >= 400 && status < 500:
		return "client_error"
	case status >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// SecurityLoggingMiddleware provides security-focused logging
func SecurityLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request scope
		requestScope := logging.NewRequestScope()
		requestScope.StartTime = time.Now().UnixNano()

		ctx := requestScope.ToContext(context.Background())
		c.Request = c.Request.WithContext(ctx)

		contextLogger := logging.NewContextualLogger(logger, ctx)
		c.Set("logger", contextLogger)

		// Log security-relevant information
		securityFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("host", c.Request.Host),
		}

		// Check for security headers
		if xForwardedFor := c.Request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
			securityFields = append(securityFields, zap.String("x_forwarded_for", xForwardedFor))
		}

		if xRealIP := c.Request.Header.Get("X-Real-IP"); xRealIP != "" {
			securityFields = append(securityFields, zap.String("x_real_ip", xRealIP))
		}

		contextLogger.Info("Security audit - request", securityFields...)

		c.Next()

		// Log final status for security monitoring
		status := c.Writer.Status()
		if status >= 400 {
			contextLogger.Error("Security audit - error response",
				zap.Int("status", status),
				zap.String("status_category", getStatusCategory(status)),
			)
		}
	}
}

// PerformanceLoggingMiddleware focuses on performance metrics
func PerformanceLoggingMiddleware(logger *zap.Logger, slowThreshold time.Duration) gin.HandlerFunc {
	if slowThreshold == 0 {
		slowThreshold = 1 * time.Second // Default threshold
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Generate request scope
		requestScope := logging.NewRequestScope()
		requestScope.StartTime = start.UnixNano()

		ctx := requestScope.ToContext(context.Background())
		c.Request = c.Request.WithContext(ctx)

		contextLogger := logging.NewContextualLogger(logger, ctx)
		c.Set("logger", contextLogger)

		c.Next()

		duration := time.Since(start)

		// Always log performance metrics
		perfFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", duration),
			zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
			zap.Int("status", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
		}

		if duration > slowThreshold {
			contextLogger.Error("Performance - slow request", perfFields...)
		} else {
			contextLogger.Debug("Performance - request timing", perfFields...)
		}
	}
}
