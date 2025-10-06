# Configuration Interface Contract

**Package**: `internal/config`  
**Purpose**: Configuration loading and validation interface

## Primary Interface

```go
// ConfigLoader defines the contract for configuration management
type ConfigLoader interface {
    // Load configuration from all sources with validation
    Load() (*Config, error)
    
    // EnableHotReload starts watching config file for changes (development only)
    EnableHotReload(callback func(*Config)) error
    
    // Validate checks configuration against business rules
    Validate(config *Config) error
}
```

## Factory Function Contract

```go
// NewConfig creates and validates a new configuration instance
// Returns error if required fields missing or validation fails
func NewConfig() (*Config, error)
```

**Input**: None (reads from environment and config.yaml)  
**Output**: 
- `*Config`: Validated configuration struct
- `error`: Validation or loading errors

**Behavior**:
- Load from environment variables (highest priority)
- Merge with config.yaml values (medium priority)  
- Apply default values (lowest priority)
- Validate all fields using go-playground/validator
- Return first validation error encountered

## Configuration Struct Contract

```go
type Config struct {
    DBDSN           string `mapstructure:"DB_DSN" validate:"required,url"`
    JWTSecret       string `mapstructure:"JWT_SECRET" validate:"required,min=32"`
    LogLevel        string `mapstructure:"LOG_LEVEL" validate:"oneof=debug info error"`
    EnableHotReload bool   `mapstructure:"ENABLE_HOT_RELOAD"`
}
```

**Field Contracts**:
- `DBDSN`: PostgreSQL connection string (required, URL format)
- `JWTSecret`: JWT signing secret (required, minimum 32 characters)
- `LogLevel`: Logging verbosity (optional, defaults to "info", enum: debug/info/error)
- `EnableHotReload`: File watching toggle (optional, defaults to false)

## Error Contracts

```go
var (
    ErrConfigNotFound     = errors.New("configuration file not found")
    ErrInvalidDSN        = errors.New("invalid database connection string")
    ErrJWTSecretTooShort = errors.New("JWT secret must be at least 32 characters")
    ErrInvalidLogLevel   = errors.New("log level must be debug, info, or error")
)
```

## Usage Contract

```go
// Standard usage pattern
config, err := NewConfig()
if err != nil {
    return fmt.Errorf("failed to load configuration: %w", err)
}

// Access validated configuration
db, err := gorm.Open(postgres.Open(config.DBDSN), &gorm.Config{})
logger, err := NewLogger(config)
```

## Hot-Reload Contract (Development Only)

```go
// Enable file watching with callback
err := config.EnableHotReload(func(newConfig *Config) {
    log.Info("Configuration reloaded", zap.String("log_level", newConfig.LogLevel))
    // Update application components with new config
})
```

**Callback Behavior**:
- Called only when config file changes and validation passes
- Receives pointer to new validated Config struct
- Application responsible for updating components with new config
- Errors in callback do not affect config reload

## Testing Contract

```go
// Test interface for mocking configuration loading
type ConfigTester interface {
    WithEnv(key, value string) ConfigTester
    WithYAML(content string) ConfigTester
    Load() (*Config, error)
    ShouldFailWith(error) bool
}
```

**Test Requirements**:
- Test valid configuration loading from environment
- Test valid configuration loading from YAML file
- Test precedence (env overrides YAML overrides defaults)
- Test validation failures for each field
- Test hot-reload functionality in development mode
- Test graceful degradation when config file missing
- Coverage requirement: >80%

---

**Contract Status**: ✅ Complete - All interfaces and behaviors defined