package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"
)

// Additional context keys for logging (RequestIDKey already defined in types.go)
const (
	userIDKey        contextKey = "user_id"
	sessionIDKey     contextKey = "session_id"
	traceIDKey       contextKey = "trace_id"
	correlationIDKey contextKey = "correlation_id"
)

// ContextualLogger provides context-aware logging functionality
// This is a concrete implementation of the ContextLogger interface defined in types.go
type ContextualLogger struct {
	logger *zap.Logger
	ctx    context.Context
}

// NewContextualLogger creates a new contextual logger
func NewContextualLogger(logger *zap.Logger, ctx context.Context) *ContextualLogger {
	return &ContextualLogger{
		logger: logger,
		ctx:    ctx,
	}
}

// WithContext returns a new logger with updated context
func (cl *ContextualLogger) WithContext(ctx context.Context) ContextLogger {
	return &ContextualLogger{
		logger: cl.logger,
		ctx:    ctx,
	}
}

// WithRequestID adds request ID to the context and returns a new logger
func (cl *ContextualLogger) WithRequestID(requestID string) ContextLogger {
	newCtx := context.WithValue(cl.ctx, RequestIDKey, requestID)
	return &ContextualLogger{
		logger: cl.logger,
		ctx:    newCtx,
	}
}

// WithFields adds structured fields and returns a new logger
func (cl *ContextualLogger) WithFields(fields map[string]interface{}) ContextLogger {
	// Store fields in context for this implementation
	// For production, consider using a more efficient approach
	newCtx := cl.ctx
	for k, v := range fields {
		newCtx = context.WithValue(newCtx, contextKey(k), v)
	}
	return &ContextualLogger{
		logger: cl.logger,
		ctx:    newCtx,
	}
}

// Debug logs a debug message with context fields
func (cl *ContextualLogger) Debug(msg string, fields ...zap.Field) {
	contextFields := cl.extractContextFields()
	allFields := append(contextFields, fields...)
	cl.logger.Debug(msg, allFields...)
}

// Info logs an info message with context fields
func (cl *ContextualLogger) Info(msg string, fields ...zap.Field) {
	contextFields := cl.extractContextFields()
	allFields := append(contextFields, fields...)
	cl.logger.Info(msg, allFields...)
}

// Error logs an error message with context fields
func (cl *ContextualLogger) Error(msg string, fields ...zap.Field) {
	contextFields := cl.extractContextFields()
	allFields := append(contextFields, fields...)
	cl.logger.Error(msg, allFields...)
}

// LogSecure logs with automatic sanitization of sensitive fields
func (cl *ContextualLogger) LogSecure(level string, message string, fields map[string]interface{}) {
	filter := NewDefaultSensitiveFieldFilter()

	sanitizedFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		sanitizedValue := filter.Sanitize(k, v)
		sanitizedFields = append(sanitizedFields, zap.Any(k, sanitizedValue))
	}

	contextFields := cl.extractContextFields()
	allFields := append(contextFields, sanitizedFields...)

	switch level {
	case "debug":
		cl.logger.Debug(message, allFields...)
	case "info":
		cl.logger.Info(message, allFields...)
	case "error":
		cl.logger.Error(message, allFields...)
	default:
		cl.logger.Info(message, allFields...)
	}
}

// extractContextFields extracts logging fields from context
func (cl *ContextualLogger) extractContextFields() []zap.Field {
	if cl.ctx == nil {
		return nil
	}

	var fields []zap.Field

	// Extract request ID
	if requestID := cl.ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			fields = append(fields, zap.String(string(RequestIDKey), id))
		}
	}

	// Extract user ID
	if userID := cl.ctx.Value(userIDKey); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			fields = append(fields, zap.String(string(userIDKey), id))
		}
	}

	// Extract session ID
	if sessionID := cl.ctx.Value(sessionIDKey); sessionID != nil {
		if id, ok := sessionID.(string); ok && id != "" {
			fields = append(fields, zap.String(string(sessionIDKey), id))
		}
	}

	// Extract trace ID
	if traceID := cl.ctx.Value(traceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			fields = append(fields, zap.String(string(traceIDKey), id))
		}
	}

	// Extract correlation ID
	if correlationID := cl.ctx.Value(correlationIDKey); correlationID != nil {
		if id, ok := correlationID.(string); ok && id != "" {
			fields = append(fields, zap.String(string(correlationIDKey), id))
		}
	}

	return fields
}

// Context helper functions

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID := ctx.Value(userIDKey); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetSessionID extracts session ID from context
func GetSessionID(ctx context.Context) string {
	if sessionID := ctx.Value(sessionIDKey); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID := ctx.Value(traceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value(correlationIDKey); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple format if random generation fails
		return fmt.Sprintf("req_%d", makeRandomInt())
	}
	return "req_" + hex.EncodeToString(bytes)
}

// GenerateTraceID generates a new trace ID
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple format if random generation fails
		return fmt.Sprintf("trace_%d", makeRandomInt())
	}
	return hex.EncodeToString(bytes)
}

// GenerateCorrelationID generates a new correlation ID
func GenerateCorrelationID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple format if random generation fails
		return fmt.Sprintf("corr_%d", makeRandomInt())
	}
	return "corr_" + hex.EncodeToString(bytes)
}

// makeRandomInt creates a random integer for fallback ID generation
func makeRandomInt() int64 {
	// Simple fallback using crypto/rand for bytes
	bytes := make([]byte, 8)
	rand.Read(bytes)
	var result int64
	for i, b := range bytes {
		result |= int64(b) << (8 * i)
	}
	if result < 0 {
		result = -result
	}
	return result
}

// LoggerFromContext creates a contextual logger from context and base logger
func LoggerFromContext(ctx context.Context, baseLogger *zap.Logger) *ContextualLogger {
	return NewContextualLogger(baseLogger, ctx)
}

// RequestScope represents a request's logging scope
type RequestScope struct {
	RequestID     string
	UserID        string
	SessionID     string
	TraceID       string
	CorrelationID string
	StartTime     int64
}

// NewRequestScope creates a new request scope with generated IDs
func NewRequestScope() *RequestScope {
	return &RequestScope{
		RequestID:     GenerateRequestID(),
		TraceID:       GenerateTraceID(),
		CorrelationID: GenerateCorrelationID(),
		StartTime:     0, // Will be set by middleware
	}
}

// ToContext converts request scope to context
func (rs *RequestScope) ToContext(ctx context.Context) context.Context {
	ctx = WithRequestID(ctx, rs.RequestID)

	if rs.UserID != "" {
		ctx = WithUserID(ctx, rs.UserID)
	}

	if rs.SessionID != "" {
		ctx = WithSessionID(ctx, rs.SessionID)
	}

	if rs.TraceID != "" {
		ctx = WithTraceID(ctx, rs.TraceID)
	}

	if rs.CorrelationID != "" {
		ctx = WithCorrelationID(ctx, rs.CorrelationID)
	}

	return ctx
}

// ToFields converts request scope to zap fields
func (rs *RequestScope) ToFields() []zap.Field {
	fields := []zap.Field{
		zap.String(string(RequestIDKey), rs.RequestID),
	}

	if rs.UserID != "" {
		fields = append(fields, zap.String(string(userIDKey), rs.UserID))
	}

	if rs.SessionID != "" {
		fields = append(fields, zap.String(string(sessionIDKey), rs.SessionID))
	}

	if rs.TraceID != "" {
		fields = append(fields, zap.String(string(traceIDKey), rs.TraceID))
	}

	if rs.CorrelationID != "" {
		fields = append(fields, zap.String(string(correlationIDKey), rs.CorrelationID))
	}

	if rs.StartTime > 0 {
		fields = append(fields, zap.Int64("start_time", rs.StartTime))
	}

	return fields
}

// Ensure ContextualLogger implements ContextLogger interface
var _ ContextLogger = (*ContextualLogger)(nil)
