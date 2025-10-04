# Data Model: Configuration and Logging Infrastructure

**Feature**: Configuration and Logging Infrastructure  
**Date**: October 4, 2025  
**Status**: Complete

## Core Entities

### Configuration Entity

**Purpose**: Represents application configuration settings loaded from multiple sources with validation

**Structure**:
```go
type Config struct {
    // Database Configuration
    DBDSN string `mapstructure:"DB_DSN" validate:"required,url"`
    
    // Authentication Configuration  
    JWTSecret string `mapstructure:"JWT_SECRET" validate:"required,min=32"`
    
    // Logging Configuration
    LogLevel string `mapstructure:"LOG_LEVEL" validate:"oneof=debug info error"`
    
    // Development Features
    EnableHotReload bool `mapstructure:"ENABLE_HOT_RELOAD"`
}
```

**Validation Rules**:
- `DBDSN`: Required PostgreSQL connection string format
- `JWTSecret`: Required, minimum 32 characters for security
- `LogLevel`: Must be one of: debug, info, error (defaults to "info")
- `EnableHotReload`: Boolean flag for development mode

**Default Values**:
- `LogLevel`: "info"
- `EnableHotReload`: false (production safe)

**Source Precedence** (highest to lowest):
1. Environment variables
2. YAML configuration file  
3. Default values

### Logger Entity

**Purpose**: Represents structured logging capability with context propagation and level filtering

**Structure**:
```go
type Logger struct {
    zapLogger *zap.Logger
    config    *Config
    level     zap.AtomicLevel
}
```

**Context Fields**:
- `request_id`: Unique identifier for request tracing
- `timestamp`: ISO 8601 formatted timestamp
- `level`: Log level (debug/info/error)
- `service`: Service name ("echoforge")
- `version`: Application version

**Output Formats**:
- **Production**: JSON structured format for machine parsing
- **Development**: Console format with colors for human readability

### Validation Rule Entity

**Purpose**: Represents validation constraints applied to configuration parameters

**Built-in Rules**:
- `required`: Field must be present and non-empty
- `url`: Valid URL format for database connections
- `min=N`: Minimum length validation for secrets
- `oneof=a b c`: Enumeration validation for log levels

**Custom Rules**:
- `postgres_dsn`: Validates PostgreSQL connection string format
- `jwt_secret_strength`: Ensures JWT secret meets security requirements

### Log Entry Entity

**Purpose**: Represents individual log messages with structured data and security sanitization

**Structure**:
```go
type LogEntry struct {
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Service   string                 `json:"service"`
    RequestID string                 `json:"request_id,omitempty"`
    Message   string                 `json:"message"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}
```

**Security Constraints**:
- Never include fields containing: `password`, `secret`, `token`, `dsn`
- All field names converted to lowercase for sanitization checks
- Sensitive values replaced with `[REDACTED]` placeholder

**Level Hierarchy**:
- `debug`: Development debugging information
- `info`: General application information (default)
- `error`: Error conditions requiring attention

## Data Relationships

### Configuration → Logger
- **Relationship**: One-to-One composition
- **Description**: Logger is created with Configuration as input
- **Factory Pattern**: `NewLogger(config *Config) (*Logger, error)`

### Logger → Log Entry  
- **Relationship**: One-to-Many creation
- **Description**: Logger creates multiple Log Entries during operation
- **Method**: `logger.Info()`, `logger.Error()`, etc.

### Configuration → Validation Rules
- **Relationship**: One-to-Many application
- **Description**: Each Configuration field has associated Validation Rules
- **Enforcement**: Applied during `NewConfig()` factory method

## State Transitions

### Configuration Lifecycle
```
[Uninitialized] → [Loading] → [Validating] → [Ready] → [Hot-Reloading]
                      ↓           ↓            ↓           ↓
                  [Load Error] [Validation Error] [Runtime] [Reload Error]
```

**States**:
- **Uninitialized**: No configuration loaded
- **Loading**: Reading from env/file sources
- **Load Error**: File not found or parsing failed
- **Validating**: Applying validation rules
- **Validation Error**: Validation rules failed
- **Ready**: Configuration valid and available
- **Runtime**: Normal operation state
- **Hot-Reloading**: Development file watching (if enabled)
- **Reload Error**: Hot-reload validation failed

### Logger Lifecycle
```
[Uninitialized] → [Configured] → [Active] → [Shutdown]
                      ↓           ↓          ↓
                 [Config Error] [Runtime] [Cleanup]
```

**States**:
- **Uninitialized**: Logger not created
- **Configured**: Logger created with valid config
- **Config Error**: Invalid logging configuration
- **Active**: Processing log entries
- **Runtime**: Normal logging operation
- **Shutdown**: Graceful cleanup and buffer flush
- **Cleanup**: Resources released

## Data Flow

### Configuration Loading Flow
```
Environment Variables → Viper → Validation → Config Struct → Application
YAML Config File     ↗       ↓               ↓
Default Values      ↗    [Merge]      [Error Handling]
```

### Logging Flow
```
Application Code → Logger Factory → Context Injection → Sanitization → Output
                      ↓                ↓                    ↓           ↓
                 Config Input    Request ID         Security Filter  JSON/Console
```

### Hot-Reload Flow (Development Only)
```
File Change → File Watcher → Debounce → Load → Validate → Update Config
                ↓              ↓         ↓       ↓           ↓
            [Event]      [Rate Limit] [Parse] [Rules]  [Hot Swap]
```

## Persistence Considerations

**No Database Persistence**: This feature handles configuration loading and logging output only. No data is persisted to the PostgreSQL database.

**File System Usage**:
- Read `configs/config.yaml` for default configuration
- Write log output to stdout/stderr (container-friendly)
- Watch `configs/config.yaml` for changes (development mode)

**Memory Storage**:
- Configuration cached in memory after initial load
- Logger instances maintained for application lifetime
- Log entries are immediately output, not stored

---

**Data Model Status**: ✅ Complete - All entities and relationships defined