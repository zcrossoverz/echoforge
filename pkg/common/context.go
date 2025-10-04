package common

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/internal/logging"
	"go.uber.org/zap"
)

// GetRequestID extracts the request ID from Gin context or underlying request context
func GetRequestID(c *gin.Context) string {
	// First try to get from Gin context (set by middleware)
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Then try to get from request context
	if requestID := logging.GetRequestID(c.Request.Context()); requestID != "" {
		return requestID
	}

	// Generate a new one if none exists
	return logging.GenerateRequestID()
}

// GetUserID extracts the user ID from Gin context or request context
func GetUserID(c *gin.Context) string {
	// First try Gin context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// Then try request context
	return logging.GetUserID(c.Request.Context())
}

// GetSessionID extracts the session ID from Gin context or request context
func GetSessionID(c *gin.Context) string {
	// First try Gin context
	if sessionID, exists := c.Get("session_id"); exists {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}

	// Then try request context
	return logging.GetSessionID(c.Request.Context())
}

// GetLogger extracts the contextual logger from Gin context
func GetLogger(c *gin.Context) *logging.ContextualLogger {
	// Try to get the contextual logger set by middleware
	if logger, exists := c.Get("logger"); exists {
		if contextLogger, ok := logger.(*logging.ContextualLogger); ok {
			return contextLogger
		}
	}

	// Fallback: try to get base logger and create contextual logger
	if baseLogger, exists := c.Get("base_logger"); exists {
		if zapLogger, ok := baseLogger.(*zap.Logger); ok {
			return logging.NewContextualLogger(zapLogger, c.Request.Context())
		}
	}

	// Last resort: create a no-op logger
	return logging.NewContextualLogger(zap.NewNop(), c.Request.Context())
}

// GetLoggerWithFields extracts logger and adds additional fields
func GetLoggerWithFields(c *gin.Context, fields map[string]interface{}) logging.ContextLogger {
	contextLogger := GetLogger(c)
	return contextLogger.WithFields(fields)
}

// InjectRequestID adds a request ID to the Gin context if it doesn't exist
func InjectRequestID(c *gin.Context) string {
	requestID := GetRequestID(c)

	// Store in Gin context
	c.Set("request_id", requestID)

	// Add to request context
	ctx := logging.WithRequestID(c.Request.Context(), requestID)
	c.Request = c.Request.WithContext(ctx)

	// Add to response headers
	c.Header("X-Request-ID", requestID)

	return requestID
}

// InjectUserContext adds user information to the request context
func InjectUserContext(c *gin.Context, userID string, sessionID ...string) {
	// Add to Gin context
	c.Set("user_id", userID)

	// Add to request context
	ctx := logging.WithUserID(c.Request.Context(), userID)

	// Add session ID if provided
	if len(sessionID) > 0 && sessionID[0] != "" {
		c.Set("session_id", sessionID[0])
		ctx = logging.WithSessionID(ctx, sessionID[0])
	}

	c.Request = c.Request.WithContext(ctx)
}

// InjectLogger adds a logger to the Gin context
func InjectLogger(c *gin.Context, logger *zap.Logger) {
	c.Set("base_logger", logger)

	// Create contextual logger
	contextLogger := logging.NewContextualLogger(logger, c.Request.Context())
	c.Set("logger", contextLogger)
}

// CreateRequestScope creates a request scope and injects it into context
func CreateRequestScope(c *gin.Context) *logging.RequestScope {
	// Check if request scope already exists
	if scope, exists := c.Get("request_scope"); exists {
		if requestScope, ok := scope.(*logging.RequestScope); ok {
			return requestScope
		}
	}

	// Create new request scope
	requestScope := logging.NewRequestScope()

	// Populate with existing context data
	if userID := GetUserID(c); userID != "" {
		requestScope.UserID = userID
	}

	if sessionID := GetSessionID(c); sessionID != "" {
		requestScope.SessionID = sessionID
	}

	// Store in Gin context
	c.Set("request_scope", requestScope)

	// Update request context
	ctx := requestScope.ToContext(c.Request.Context())
	c.Request = c.Request.WithContext(ctx)

	return requestScope
}

// LogRequest logs basic request information using the contextual logger
func LogRequest(c *gin.Context, message string, fields ...zap.Field) {
	logger := GetLogger(c)

	// Add basic request fields
	requestFields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
	}

	allFields := append(requestFields, fields...)
	logger.Info(message, allFields...)
}

// LogError logs an error with request context
func LogError(c *gin.Context, err error, message string, fields ...zap.Field) {
	logger := GetLogger(c)

	// Add basic request fields
	requestFields := []zap.Field{
		zap.Error(err),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
	}

	allFields := append(requestFields, fields...)
	logger.Error(message, allFields...)
}

// LogSecure logs with automatic sanitization of sensitive data
func LogSecure(c *gin.Context, level string, message string, fields map[string]interface{}) {
	logger := GetLogger(c)

	// Add request context to fields
	enhancedFields := make(map[string]interface{})
	for k, v := range fields {
		enhancedFields[k] = v
	}

	enhancedFields["method"] = c.Request.Method
	enhancedFields["path"] = c.Request.URL.Path
	enhancedFields["remote_addr"] = c.ClientIP()

	logger.LogSecure(level, message, enhancedFields)
}

// Handler helper functions for common patterns

// WithLogging is a handler wrapper that ensures logging context is available
func WithLogging(handler gin.HandlerFunc, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure request ID exists
		InjectRequestID(c)

		// Ensure logger is available
		if _, exists := c.Get("logger"); !exists {
			InjectLogger(c, logger)
		}

		// Call the actual handler
		handler(c)
	}
}

// WithRequestScope is a handler wrapper that ensures request scope is available
func WithRequestScope(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure request scope exists
		CreateRequestScope(c)

		// Call the actual handler
		handler(c)
	}
}

// AuthContext represents authenticated user context
type AuthContext struct {
	UserID    string
	SessionID string
	Username  string
	Roles     []string
}

// InjectAuthContext adds authentication context to the request
func InjectAuthContext(c *gin.Context, authCtx *AuthContext) {
	// Store in Gin context
	c.Set("auth_context", authCtx)
	c.Set("user_id", authCtx.UserID)

	if authCtx.SessionID != "" {
		c.Set("session_id", authCtx.SessionID)
	}

	// Add to request context
	ctx := logging.WithUserID(c.Request.Context(), authCtx.UserID)

	if authCtx.SessionID != "" {
		ctx = logging.WithSessionID(ctx, authCtx.SessionID)
	}

	c.Request = c.Request.WithContext(ctx)

	// Log authentication event
	logger := GetLogger(c)
	logger.Info("User authenticated",
		zap.String("user_id", authCtx.UserID),
		zap.String("username", authCtx.Username),
		zap.Strings("roles", authCtx.Roles),
	)
}

// GetAuthContext extracts authentication context from Gin context
func GetAuthContext(c *gin.Context) *AuthContext {
	if authCtx, exists := c.Get("auth_context"); exists {
		if auth, ok := authCtx.(*AuthContext); ok {
			return auth
		}
	}
	return nil
}

// RequireAuth is a middleware that requires authentication context
func RequireAuth(c *gin.Context) {
	authCtx := GetAuthContext(c)
	if authCtx == nil {
		logger := GetLogger(c)
		logger.Error("Authentication required but not found")

		c.JSON(401, gin.H{"error": "Authentication required"})
		c.Abort()
		return
	}

	c.Next()
}

// LogWithTrace adds trace information to log entries
func LogWithTrace(c *gin.Context, level string, message string, fields ...zap.Field) {
	logger := GetLogger(c)

	// Add trace ID if available
	traceID := logging.GetTraceID(c.Request.Context())
	if traceID == "" {
		traceID = logging.GenerateTraceID()
		ctx := logging.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
	}

	traceFields := append(fields, zap.String("trace_id", traceID))

	switch level {
	case "debug":
		logger.Debug(message, traceFields...)
	case "info":
		logger.Info(message, traceFields...)
	case "error":
		logger.Error(message, traceFields...)
	default:
		logger.Info(message, traceFields...)
	}
}

// Performance tracking helpers

// StartTimer creates a performance timer for request operations
func StartTimer(c *gin.Context, operation string) func() {
	start := time.Now()

	return func() {
		duration := time.Since(start)
		logger := GetLogger(c)
		logger.Info("Operation completed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}
}

// TrackOperation is a helper to track operation performance
func TrackOperation(c *gin.Context, operation string, fn func() error) error {
	timer := StartTimer(c, operation)
	defer timer()

	logger := GetLogger(c)
	logger.Debug("Operation started", zap.String("operation", operation))

	err := fn()
	if err != nil {
		logger.Error("Operation failed",
			zap.String("operation", operation),
			zap.Error(err),
		)
	}

	return err
}

// Correlation ID helpers for distributed tracing

// GetOrCreateCorrelationID gets or creates a correlation ID for the request
func GetOrCreateCorrelationID(c *gin.Context) string {
	// Check headers first
	if corrID := c.GetHeader("X-Correlation-ID"); corrID != "" {
		return corrID
	}

	// Check context
	if corrID := logging.GetCorrelationID(c.Request.Context()); corrID != "" {
		return corrID
	}

	// Generate new one
	corrID := logging.GenerateCorrelationID()

	// Add to context
	ctx := logging.WithCorrelationID(c.Request.Context(), corrID)
	c.Request = c.Request.WithContext(ctx)

	// Add to response headers
	c.Header("X-Correlation-ID", corrID)

	return corrID
}

// PropagateCorrelationID ensures correlation ID is propagated in outbound requests
func PropagateCorrelationID(c *gin.Context, headers map[string]string) {
	corrID := GetOrCreateCorrelationID(c)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["X-Correlation-ID"] = corrID
}
