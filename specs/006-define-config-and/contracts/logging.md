# Logging Interface Contract

**Package**: `internal/logging`  
**Purpose**: Structured logging with context propagation and security sanitization

## Primary Interface

```go
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
```

## Factory Function Contract

```go
// NewLogger creates a new logger instance configured for the environment
func NewLogger(config *Config) (*zap.Logger, error)
```

**Input**: 
- `*Config`: Configuration containing LogLevel and environment settings

**Output**:
- `*zap.Logger`: Configured Zap logger instance
- `error`: Configuration or initialization errors

**Behavior**:
- Create production config (JSON) for deployed environments
- Create development config (console) for local development
- Set log level based on config.LogLevel
- Configure security sanitization for sensitive fields
- Enable context propagation support

## Logger Configuration Contract

```go
type LoggerConfig struct {
    Level       string // debug, info, error
    Format      string // json, console  
    Output      string // stdout, stderr, file path
    Sampling    bool   // Enable sampling for high-volume scenarios
    Development bool   // Enable development features (color, verbose)
}
```

## Context Integration Contract

```go
// Context key for request ID propagation
type contextKey string
const RequestIDKey contextKey = "request_id"

// Middleware contract for Gin integration
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := generateRequestID()
        c.Set(string(RequestIDKey), requestID)
        
        // Log request start
        logger.Info("Request started",
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
        )
        
        c.Next()
        
        // Log request completion
        logger.Info("Request completed",
            zap.String("request_id", requestID),
            zap.Int("status", c.Writer.Status()),
        )
    }
}
```

## Security Sanitization Contract

```go
// SensitiveFieldFilter defines fields that must be sanitized
type SensitiveFieldFilter interface {
    IsSensitive(fieldName string) bool
    Sanitize(fieldName string, value interface{}) interface{}
}

// Default sensitive field patterns
var SensitiveFields = []string{
    "password", "secret", "token", "dsn", "key", "auth", "credential"
}
```

**Sanitization Rules**:
- Field names checked case-insensitively
- Sensitive values replaced with `[REDACTED]`
- Structured data recursively sanitized
- URL query parameters and form data sanitized
- Database connection strings completely redacted

## Log Entry Format Contract

### Production Format (JSON)
```json
{
    "timestamp": "2025-10-04T10:30:45.123Z",
    "level": "info",
    "service": "echoforge",
    "version": "1.0.0",
    "request_id": "req_abc123def456",
    "message": "User authentication successful",
    "fields": {
        "user_id": "user_789",
        "email": "user@example.com",
        "duration_ms": 45
    }
}
```

### Development Format (Console)
```
2025-10-04 10:30:45 INFO [req_abc123def456] User authentication successful user_id=user_789 email=user@example.com duration_ms=45
```

## Error Handling Contract

```go
var (
    ErrInvalidLogLevel   = errors.New("invalid log level specified")
    ErrLoggerNotConfigured = errors.New("logger not properly configured")
    ErrContextMissing    = errors.New("required context missing from request")
)
```

## Performance Contract

**Requirements**:
- Support minimum 1000 log entries per second
- Context propagation overhead <1ms per request
- Memory usage <50MB for logging infrastructure
- Buffer management for burst scenarios

**Sampling Strategy**:
```go
// High-volume logging sampling configuration
type SamplingConfig struct {
    Initial    int // Log first N entries per second
    Thereafter int // Then log every Nth entry
}
```

## Wire Dependency Injection Contract

```go
// Wire provider set for dependency injection
var LoggingSet = wire.NewSet(
    NewLogger,
    NewSensitiveFieldFilter,
    wire.Bind(new(ContextLogger), new(*zap.Logger)),
)
```

## Testing Contract

```go
// Test interface for verifying log output
type LogTester interface {
    CaptureOutput() LogTester
    WithLevel(level string) LogTester
    WithContext(ctx context.Context) LogTester
    ExpectMessage(message string) LogTester
    ExpectField(key string, value interface{}) LogTester
    ExpectNoSensitiveData() LogTester
    Verify() error
}
```

**Test Requirements**:
- Test log level filtering (debug < info < error)
- Test JSON format structure in production mode
- Test console format readability in development mode
- Test context propagation through request lifecycle
- Test sensitive data sanitization for all field types
- Test performance under high-volume scenarios
- Test graceful degradation when context missing
- Coverage requirement: >80%

## Integration with Existing Systems

### Gin HTTP Framework
```go
// Logger middleware integration
r := gin.New()
r.Use(LoggingMiddleware(logger))

// Handler-level logging
func userHandler(c *gin.Context) {
    logger := c.MustGet("logger").(*zap.Logger)
    logger.Info("Processing user request")
}
```

### GORM Database Integration
```go
// Database query logging (sanitized)
logger.Debug("Database query executed",
    zap.String("table", "users"),
    zap.Duration("duration", queryTime),
    // Note: SQL query and parameters sanitized automatically
)
```

---

**Contract Status**: ✅ Complete - All interfaces and integration points defined