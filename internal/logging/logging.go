package logging

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config interface to avoid circular import
type Config interface {
	GetLogLevel() string
	IsDevelopment() bool
}

// SimpleConfig implements Config interface for the logger
type SimpleConfig struct {
	LogLevel    string
	Development bool
}

func (c *SimpleConfig) GetLogLevel() string {
	return c.LogLevel
}

func (c *SimpleConfig) IsDevelopment() bool {
	return c.Development
}

// NewLogger creates a new logger instance configured for the environment
func NewLogger(config Config) (*zap.Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logLevel := config.GetLogLevel()
	isDevelopment := config.IsDevelopment()

	// Parse log level
	var zapLevel zapcore.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, or error)", logLevel)
	}

	var zapConfig zap.Config

	if isDevelopment {
		// Development configuration: console output with colors
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Production configuration: JSON output
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.MessageKey = "message"
		zapConfig.EncoderConfig.LevelKey = "level"
	}

	// Set the log level
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	// Configure output
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	// Build the logger
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// Add default fields
	logger = logger.With(
		zap.String("service", "echoforge"),
		zap.String("version", "1.0.0"),
	)

	return logger, nil
}

// NewLoggerFromLevel creates a logger with just a log level (convenience function)
func NewLoggerFromLevel(logLevel string) (*zap.Logger, error) {
	config := &SimpleConfig{
		LogLevel:    logLevel,
		Development: logLevel == "debug" || os.Getenv("GIN_MODE") != "release",
	}
	return NewLogger(config)
}

// NewContextLogger creates a context-aware logger wrapper
func NewContextLogger(logger *zap.Logger) ContextLogger {
	filter := NewDefaultSensitiveFieldFilter()
	return NewZapContextLogger(logger, filter)
}

// MustNewLogger creates a logger or panics on error (for initialization)
func MustNewLogger(config Config) *zap.Logger {
	logger, err := NewLogger(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// LoggerBuilder provides a fluent interface for logger configuration
type LoggerBuilder struct {
	level       string
	development bool
	service     string
	version     string
	sampling    *SamplingConfig
	fields      map[string]interface{}
}

// NewLoggerBuilder creates a new logger builder
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:   "info",
		service: "echoforge",
		version: "1.0.0",
		fields:  make(map[string]interface{}),
	}
}

// WithLevel sets the log level
func (b *LoggerBuilder) WithLevel(level string) *LoggerBuilder {
	b.level = level
	return b
}

// WithDevelopment enables development mode
func (b *LoggerBuilder) WithDevelopment(dev bool) *LoggerBuilder {
	b.development = dev
	return b
}

// WithService sets the service name
func (b *LoggerBuilder) WithService(service string) *LoggerBuilder {
	b.service = service
	return b
}

// WithVersion sets the service version
func (b *LoggerBuilder) WithVersion(version string) *LoggerBuilder {
	b.version = version
	return b
}

// WithSampling configures log sampling
func (b *LoggerBuilder) WithSampling(config SamplingConfig) *LoggerBuilder {
	b.sampling = &config
	return b
}

// WithField adds a default field to all log entries
func (b *LoggerBuilder) WithField(key string, value interface{}) *LoggerBuilder {
	b.fields[key] = value
	return b
}

// Build creates the logger with the configured options
func (b *LoggerBuilder) Build() (*zap.Logger, error) {
	config := &SimpleConfig{
		LogLevel:    b.level,
		Development: b.development,
	}

	logger, err := NewLogger(config)
	if err != nil {
		return nil, err
	}

	// Add custom fields
	fields := make([]zap.Field, 0, len(b.fields)+2)
	fields = append(fields, zap.String("service", b.service))
	fields = append(fields, zap.String("version", b.version))

	for k, v := range b.fields {
		fields = append(fields, zap.Any(k, v))
	}

	if len(fields) > 0 {
		logger = logger.With(fields...)
	}

	// Configure sampling if specified
	if b.sampling != nil {
		// Sampling configuration would be applied here
		// For now, just return the logger as-is
	}

	return logger, nil
}

// CreateDevelopmentLogger creates a pre-configured development logger
func CreateDevelopmentLogger() (*zap.Logger, error) {
	return NewLoggerBuilder().
		WithLevel("debug").
		WithDevelopment(true).
		Build()
}

// CreateProductionLogger creates a pre-configured production logger
func CreateProductionLogger() (*zap.Logger, error) {
	return NewLoggerBuilder().
		WithLevel("info").
		WithDevelopment(false).
		Build()
}
