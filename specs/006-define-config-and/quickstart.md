# Quickstart: Configuration and Logging Infrastructure

**Feature**: Configuration and Logging Infrastructure  
**Purpose**: End-to-end validation of configuration loading and structured logging  
**Time**: ~5 minutes setup + testing

## Prerequisites

- Go 1.25+ installed
- echoforge repository cloned
- Feature branch `006-define-config-and` checked out

## Quick Setup

### 1. Install Dependencies
```bash
cd echoforge
go mod tidy
```

**Expected new dependencies**:
- `github.com/spf13/viper v1.19.0`
- `go.uber.org/zap v1.27.0`
- `github.com/go-playground/validator/v10`

### 2. Create Default Configuration
```bash
# Create configs directory if not exists
mkdir -p configs

# Create default config.yaml
cat > configs/config.yaml << EOF
DB_DSN: "postgres://user:pass@localhost:5432/echoforge_dev?sslmode=disable"
JWT_SECRET: "your-super-secret-jwt-key-at-least-32-chars-long"
LOG_LEVEL: "info"
ENABLE_HOT_RELOAD: true
EOF
```

### 3. Set Environment Variables (Optional)
```bash
# Override config with environment variables
export DB_DSN="postgres://user:pass@localhost:5432/echoforge_prod?sslmode=disable"
export JWT_SECRET="production-jwt-secret-key-32-characters-minimum"
export LOG_LEVEL="error"
```

## Core Usage Examples

### 1. Basic Configuration Loading
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/zcrossoverz/echoforge/internal/config"
)

func main() {
    // Load configuration with validation
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    fmt.Printf("Database: %s\n", cfg.DBDSN)
    fmt.Printf("Log Level: %s\n", cfg.LogLevel)
    fmt.Printf("Hot Reload: %v\n", cfg.EnableHotReload)
}
```

### 2. Basic Structured Logging
```go
package main

import (
    "context"
    "log"
    
    "go.uber.org/zap"
    "github.com/zcrossoverz/echoforge/internal/config"
    "github.com/zcrossoverz/echoforge/internal/logging"
)

func main() {
    // Load config and create logger
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatalf("Config error: %v", err)
    }
    
    logger, err := logging.NewLogger(cfg)
    if err != nil {
        log.Fatalf("Logger error: %v", err)
    }
    defer logger.Sync()
    
    // Structured logging with context
    ctx := context.WithValue(context.Background(), "request_id", "req_123")
    
    logger.Info("Application started",
        zap.String("service", "echoforge"),
        zap.String("version", "1.0.0"),
    )
    
    logger.Debug("Debug information",
        zap.String("request_id", ctx.Value("request_id").(string)),
        zap.Int("user_count", 42),
    )
    
    logger.Error("Error occurred",
        zap.Error(fmt.Errorf("example error")),
        zap.String("component", "user_service"),
    )
}
```

### 3. Security Sanitization Demo
```go
package main

import (
    "log"
    
    "go.uber.org/zap"
    "github.com/zcrossoverz/echoforge/internal/config"
    "github.com/zcrossoverz/echoforge/internal/logging"
)

func main() {
    cfg, _ := config.NewConfig()
    logger, _ := logging.NewLogger(cfg)
    defer logger.Sync()
    
    // These sensitive fields will be automatically sanitized
    logger.Info("User login attempt",
        zap.String("email", "user@example.com"),         // Safe: logged
        zap.String("password", "secret123"),             // Unsafe: [REDACTED]
        zap.String("jwt_token", "eyJ0eXAiOiJKV1Q..."),   // Unsafe: [REDACTED]
        zap.String("db_dsn", cfg.DBDSN),                 // Unsafe: [REDACTED]
    )
}
```

## Testing Scenarios

### 1. Configuration Validation Tests
```bash
# Test invalid database DSN
export DB_DSN="invalid-url"
go run main.go
# Expected: Error message about invalid URL format

# Test JWT secret too short
export JWT_SECRET="short"
go run main.go  
# Expected: Error message about minimum 32 characters

# Test invalid log level
export LOG_LEVEL="verbose"
go run main.go
# Expected: Error message about valid levels (debug/info/error)
```

### 2. Hot-Reload Testing (Development)
```bash
# Terminal 1: Start application with hot-reload
export ENABLE_HOT_RELOAD=true
go run main.go

# Terminal 2: Modify config.yaml
echo 'LOG_LEVEL: "debug"' >> configs/config.yaml

# Expected: Terminal 1 shows "Configuration reloaded" message
# New log entries use debug level
```

### 3. Log Level Filtering
```bash
# Test debug level (shows all logs)
export LOG_LEVEL="debug"
go run main.go

# Test info level (shows info and error)
export LOG_LEVEL="info"
go run main.go

# Test error level (shows only errors)
export LOG_LEVEL="error"
go run main.go
```

### 4. JSON vs Console Output
```bash
# Development mode (console output with colors)
go run main.go

# Production mode (JSON structured output)
export GIN_MODE=release
go run main.go
```

## Integration with Existing Code

### 1. Gin Middleware Integration
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/zcrossoverz/echoforge/internal/config"
    "github.com/zcrossoverz/echoforge/internal/logging"
)

func main() {
    cfg, _ := config.NewConfig()
    logger, _ := logging.NewLogger(cfg)
    
    r := gin.New()
    r.Use(logging.GinMiddleware(logger))
    
    r.GET("/health", func(c *gin.Context) {
        // Logger automatically includes request ID from middleware
        logger.Info("Health check requested")
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    r.Run(":8080")
}
```

### 2. GORM Integration
```go
package main

import (
    "gorm.io/driver/postgres"  
    "gorm.io/gorm"
    "github.com/zcrossoverz/echoforge/internal/config"
)

func main() {
    cfg, _ := config.NewConfig()
    
    // Use validated DB_DSN from config
    db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
    if err != nil {
        logger.Error("Database connection failed", zap.Error(err))
        return
    }
    
    logger.Info("Database connected successfully")
}
```

## Expected Results

### 1. Successful Configuration Load
```
2025-10-04T10:30:45.123Z INFO Configuration loaded successfully
2025-10-04T10:30:45.124Z INFO Logger initialized log_level=info format=json
```

### 2. Structured JSON Output (Production)
```json
{"timestamp":"2025-10-04T10:30:45.123Z","level":"info","service":"echoforge","message":"User authentication successful","fields":{"user_id":"123","email":"user@example.com","password":"[REDACTED]"}}
```

### 3. Console Output (Development)
```
2025-10-04 10:30:45 INFO [req_abc123] User authentication successful user_id=123 email=user@example.com password=[REDACTED]
```

### 4. Configuration Validation Error
```
2025-10-04T10:30:45.123Z ERROR Configuration validation failed error="Key: 'Config.JWTSecret' Error:Field validation for 'JWTSecret' failed on the 'min' tag"
```

## Troubleshooting

### Common Issues

1. **"Configuration file not found"**
   - Ensure `configs/config.yaml` exists
   - Check file permissions are readable
   - Verify working directory is repository root

2. **"Invalid log level"**
   - Use only: `debug`, `info`, or `error`
   - Check for typos in environment variables
   - Verify YAML syntax in config file

3. **"JWT secret too short"**
   - Minimum 32 characters required
   - Use strong, random secret for production
   - Check environment variable is set correctly

4. **Hot-reload not working**
   - Ensure `ENABLE_HOT_RELOAD=true`
   - Check file permissions for config.yaml
   - Verify file watcher has access to configs directory

### Performance Issues

1. **High memory usage**
   - Check log level isn't set to debug in production
   - Verify log sampling is enabled for high-volume scenarios
   - Monitor for log entry accumulation

2. **Slow configuration loading**
   - Verify network access to database for DSN validation
   - Check YAML file size (should be minimal)
   - Consider caching configuration in memory

## Success Criteria

✅ **Configuration loads successfully** from env and YAML  
✅ **Validation catches invalid values** with clear error messages  
✅ **Structured logging outputs** JSON (prod) and console (dev) formats  
✅ **Sensitive data is sanitized** in all log output  
✅ **Context propagation works** through request lifecycle  
✅ **Hot-reload functions** in development mode  
✅ **Performance meets targets**: 1000+ logs/sec, <5s config load  
✅ **Integration works** with existing Gin and GORM components

---

**Quickstart Status**: ✅ Complete - Ready for implementation validation