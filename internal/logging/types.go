package logging

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

// LogEntry represents individual log messages with structured data and security sanitization
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	RequestID string                 `json:"request_id,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// LoggerConfig represents logger configuration options
type LoggerConfig struct {
	Level       string // debug, info, error
	Format      string // json, console
	Output      string // stdout, stderr, file path
	Sampling    bool   // Enable sampling for high-volume scenarios
	Development bool   // Enable development features (color, verbose)
}

// SensitiveFieldFilter defines fields that must be sanitized
type SensitiveFieldFilter interface {
	IsSensitive(fieldName string) bool
	Sanitize(fieldName string, value interface{}) interface{}
}

// DefaultSensitiveFieldFilter implements SensitiveFieldFilter
type DefaultSensitiveFieldFilter struct {
	sensitivePatterns []string
}

// NewDefaultSensitiveFieldFilter creates a new filter with default sensitive field patterns
func NewDefaultSensitiveFieldFilter() *DefaultSensitiveFieldFilter {
	return &DefaultSensitiveFieldFilter{
		sensitivePatterns: []string{
			"password", "secret", "token", "dsn", "key", "auth", "credential",
		},
	}
}

// IsSensitive checks if a field name matches sensitive patterns
func (f *DefaultSensitiveFieldFilter) IsSensitive(fieldName string) bool {
	lowerFieldName := strings.ToLower(fieldName)

	for _, pattern := range f.sensitivePatterns {
		if strings.Contains(lowerFieldName, pattern) {
			return true
		}
	}

	return false
}

// Sanitize replaces sensitive values with [REDACTED]
func (f *DefaultSensitiveFieldFilter) Sanitize(fieldName string, value interface{}) interface{} {
	if f.IsSensitive(fieldName) {
		return "[REDACTED]"
	}
	return value
}

// ContextLogger defines the contract for context-aware structured logging
type ContextLogger interface {
	// Context-aware logging methods
	WithContext(ctx context.Context) ContextLogger
	WithRequestID(requestID string) ContextLogger
	WithFields(fields map[string]interface{}) ContextLogger

	// Level-specific logging
	Debug(message string, fields ...zap.Field)
	Info(message string, fields ...zap.Field)
	Error(message string, fields ...zap.Field)

	// Structured logging with automatic sanitization
	LogSecure(level string, message string, fields map[string]interface{})
}

// RequestIDKey is the context key for request ID propagation
type contextKey string

const RequestIDKey contextKey = "request_id"

// ZapContextLogger wraps zap.Logger with context-aware functionality
type ZapContextLogger struct {
	logger *zap.Logger
	filter SensitiveFieldFilter
	ctx    context.Context
	fields map[string]interface{}
}

// NewZapContextLogger creates a new context-aware logger wrapper
func NewZapContextLogger(logger *zap.Logger, filter SensitiveFieldFilter) *ZapContextLogger {
	return &ZapContextLogger{
		logger: logger,
		filter: filter,
		ctx:    context.Background(),
		fields: make(map[string]interface{}),
	}
}

// WithContext returns a new logger with the given context
func (l *ZapContextLogger) WithContext(ctx context.Context) ContextLogger {
	return &ZapContextLogger{
		logger: l.logger,
		filter: l.filter,
		ctx:    ctx,
		fields: l.fields,
	}
}

// WithRequestID returns a new logger with the request ID set
func (l *ZapContextLogger) WithRequestID(requestID string) ContextLogger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields["request_id"] = requestID

	return &ZapContextLogger{
		logger: l.logger,
		filter: l.filter,
		ctx:    l.ctx,
		fields: newFields,
	}
}

// WithFields returns a new logger with additional fields
func (l *ZapContextLogger) WithFields(fields map[string]interface{}) ContextLogger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &ZapContextLogger{
		logger: l.logger,
		filter: l.filter,
		ctx:    l.ctx,
		fields: newFields,
	}
}

// Debug logs a debug message with context and sanitization
func (l *ZapContextLogger) Debug(message string, fields ...zap.Field) {
	l.logWithContext(l.logger.Debug, message, fields...)
}

// Info logs an info message with context and sanitization
func (l *ZapContextLogger) Info(message string, fields ...zap.Field) {
	l.logWithContext(l.logger.Info, message, fields...)
}

// Error logs an error message with context and sanitization
func (l *ZapContextLogger) Error(message string, fields ...zap.Field) {
	l.logWithContext(l.logger.Error, message, fields...)
}

// LogSecure logs with automatic sanitization of sensitive fields
func (l *ZapContextLogger) LogSecure(level string, message string, fields map[string]interface{}) {
	// Sanitize fields
	sanitizedFields := make(map[string]interface{})
	for k, v := range fields {
		sanitizedFields[k] = l.filter.Sanitize(k, v)
	}

	// Add context fields
	for k, v := range l.fields {
		if _, exists := sanitizedFields[k]; !exists {
			sanitizedFields[k] = l.filter.Sanitize(k, v)
		}
	}

	// Add request ID from context if available
	if requestID := l.ctx.Value(RequestIDKey); requestID != nil {
		sanitizedFields["request_id"] = requestID
	}

	// Convert to zap fields
	zapFields := make([]zap.Field, 0, len(sanitizedFields))
	for k, v := range sanitizedFields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	// Log at appropriate level
	switch strings.ToLower(level) {
	case "debug":
		l.logger.Debug(message, zapFields...)
	case "info":
		l.logger.Info(message, zapFields...)
	case "error":
		l.logger.Error(message, zapFields...)
	default:
		l.logger.Info(message, zapFields...)
	}
}

// logWithContext is a helper to add context information to log calls
func (l *ZapContextLogger) logWithContext(logFunc func(string, ...zap.Field), message string, fields ...zap.Field) {
	// Add context fields
	contextFields := make([]zap.Field, 0, len(l.fields)+1)

	// Add request ID from context
	if requestID := l.ctx.Value(RequestIDKey); requestID != nil {
		contextFields = append(contextFields, zap.Any("request_id", requestID))
	}

	// Add other context fields with sanitization
	for k, v := range l.fields {
		sanitizedValue := l.filter.Sanitize(k, v)
		contextFields = append(contextFields, zap.Any(k, sanitizedValue))
	}

	// Combine with provided fields (also sanitize them)
	allFields := append(contextFields, fields...)

	// Call the actual log function
	logFunc(message, allFields...)
}

// SamplingConfig represents configuration for high-volume logging sampling
type SamplingConfig struct {
	Initial    int // Log first N entries per second
	Thereafter int // Then log every Nth entry
}

// LoggingMiddlewareConfig represents configuration for HTTP logging middleware
type LoggingMiddlewareConfig struct {
	IncludeRequestBody  bool
	IncludeResponseBody bool
	MaxBodySize         int
	SensitiveHeaders    []string
}

// DefaultLoggingMiddlewareConfig returns default middleware configuration
func DefaultLoggingMiddlewareConfig() LoggingMiddlewareConfig {
	return LoggingMiddlewareConfig{
		IncludeRequestBody:  false, // Don't log request body by default (may contain sensitive data)
		IncludeResponseBody: false, // Don't log response body by default
		MaxBodySize:         1024,  // 1KB max body size if enabled
		SensitiveHeaders: []string{
			"authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token",
		},
	}
}

// LogLevel represents logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case ErrorLevel:
		return "error"
	default:
		return "info"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
