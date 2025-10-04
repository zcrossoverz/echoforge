# Research: Configuration and Logging Infrastructure

**Feature**: Configuration and Logging Infrastructure  
**Date**: October 4, 2025  
**Status**: Complete

## Research Questions

### 1. Viper v1.19.0 Configuration Loading Patterns

**Decision**: Use Viper with hierarchical configuration loading (env → yaml → defaults)

**Rationale**: 
- Viper provides mature support for multiple configuration sources with clear precedence
- Built-in environment variable binding with automatic type conversion
- YAML file watching for hot-reload capability
- Validation integration with go-playground/validator/v10

**Alternatives Considered**:
- Direct env parsing: Too basic, no file support or validation
- Custom config system: Reinventing the wheel, more maintenance overhead

**Implementation Approach**:
```go
// Precedence: Environment → YAML → Defaults
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath("./configs")
viper.AutomaticEnv()
```

### 2. Zap v1.27.0 Structured Logging Best Practices

**Decision**: Use Zap with production and development configurations

**Rationale**:
- JSON structured output for production (machine readable)
- Console output for development (human readable)
- High performance (faster than stdlib log)
- Context propagation support for request tracing

**Alternatives Considered**:
- Standard library log: No structured logging, poor performance
- Logrus: Slower than Zap, less flexible
- Slog (Go 1.21+): Good alternative but Zap has more ecosystem support

**Implementation Approach**:
```go
// Production: JSON structured logging
// Development: Console logging with color
// Context-aware with request ID propagation
```

### 3. OWASP-Compliant Logging Security

**Decision**: Implement sensitive data sanitization with allowlist approach

**Rationale**:
- Prevent accidental logging of DB_DSN connection strings
- Block JWT_SECRET from appearing in any log output  
- Use field-level sanitization rather than string replacement
- Allowlist approach: only log explicitly safe fields

**Alternatives Considered**:
- Regex-based scrubbing: Brittle, can miss patterns
- Blocklist approach: Easy to forget sensitive fields
- No logging security: Unacceptable security risk

**Implementation Approach**:
- Custom Zap encoder with sensitive field detection
- Structured logging with explicit field tagging
- Unit tests to verify no secrets in log output

### 4. Context-Aware Request Tracing

**Decision**: Use context.Context with request ID propagation through middleware

**Rationale**:
- Standard Go pattern for request-scoped data
- Compatible with existing Gin middleware
- Enables tracing of user actions across service layers
- Minimal performance overhead

**Alternatives Considered**:
- Thread-local storage: Not idiomatic in Go
- Global request ID: Race conditions, not concurrent-safe
- No request tracking: Poor debugging experience

**Implementation Approach**:
```go
// Gin middleware adds request ID to context
// Logger factory accepts context and extracts request ID
// All log calls include request context
```

### 5. Hot-Reload Implementation for Development

**Decision**: Use Viper's file watching with debounced reload

**Rationale**:
- Viper provides built-in file system watching
- Debouncing prevents multiple reloads during file saves
- Development-only feature (disabled in production)
- Graceful fallback if reload fails

**Alternatives Considered**:
- Manual file polling: Higher resource usage
- External file watcher: Additional dependency
- No hot-reload: Poor developer experience

**Implementation Approach**:
```go
// Use viper.WatchConfig() with callback
// Debounce rapid file changes (500ms delay)
// Validate config before applying changes
// Log reload success/failure
```

### 6. Go-Playground Validator Integration

**Decision**: Use struct tags with custom validation rules for config

**Rationale**:
- Declarative validation with struct tags
- Built-in validators for common patterns (url, email, etc.)
- Custom validators for domain-specific rules (JWT secret length)
- Clear error messages for configuration issues

**Alternatives Considered**:
- Manual validation: Verbose, error-prone
- Different validation library: Less ecosystem support
- No validation: Poor error handling experience

**Implementation Approach**:
```go
type Config struct {
    DBDSN     string `validate:"required,url"`
    JWTSecret string `validate:"required,min=32"`
    LogLevel  string `validate:"oneof=debug info error"`
}
```

## Technical Decisions Summary

| Component | Technology | Version | Rationale |
|-----------|------------|---------|-----------|
| Config Loading | Viper | v1.19.0 | Multi-source support, hot-reload, mature |
| Validation | go-playground/validator | v10 | Struct tags, custom rules, clear errors |
| Logging | Zap | v1.27.0 | Performance, structured output, context support |
| Security | Custom sanitizer | N/A | OWASP compliance, sensitive data protection |

## Performance Considerations

- **Config Loading**: One-time startup cost, cached in memory
- **Hot-Reload**: Development-only, file watching has minimal overhead
- **Logging**: Zap benchmarks show >1M logs/second capability
- **Memory**: Estimated <10MB overhead for config + logging infrastructure

## Security Considerations

- **Sensitive Data**: Never log DB_DSN, JWT_SECRET, or password fields
- **Input Validation**: All config values validated before use
- **File Permissions**: Config files should be readable only by application user
- **Environment Variables**: Preferred for secrets in production deployments

---

**Research Status**: ✅ Complete - All technical decisions finalized