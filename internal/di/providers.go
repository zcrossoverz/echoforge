package di

import (
	"github.com/google/wire"
	"github.com/zcrossoverz/echoforge/internal/config"
	"github.com/zcrossoverz/echoforge/internal/logging"
	"go.uber.org/zap"
)

// ConfigProvider provides a configured Config instance
func ConfigProvider() (*config.Config, error) {
	return config.NewConfig()
}

// ConfigWithWatcherProvider provides a Config with optional hot-reload watcher
func ConfigWithWatcherProvider() (*config.ConfigWithWatcher, error) {
	return config.NewConfigWithHotReload()
}

// LoggerProvider provides a configured zap.Logger instance
func LoggerProvider(cfg *config.Config) (*zap.Logger, error) {
	loggerConfig := &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: false,
	}

	return logging.NewLogger(loggerConfig)
}

// ContextLoggerProvider provides a context-aware logger
func ContextLoggerProvider(logger *zap.Logger) logging.ContextLogger {
	// Create a contextual logger with a background context initially
	// The actual context will be provided when used in handlers/middleware
	return logging.NewContextualLogger(logger, nil)
}

// DevelopmentLoggerProvider provides a development-optimized logger
func DevelopmentLoggerProvider(cfg *config.Config) (*zap.Logger, error) {
	loggerConfig := &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: true, // Enable development features
	}

	return logging.NewLogger(loggerConfig)
}

// ProductionLoggerProvider provides a production-optimized logger with sanitization
func ProductionLoggerProvider(cfg *config.Config) (*zap.Logger, error) {
	loggerConfig := &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: false, // Disable development features
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		return nil, err
	}

	// Wrap with sanitization for production
	return logging.CreateSanitizingLogger(logger), nil
}

// SensitiveFieldFilterProvider provides the default sensitive field filter
func SensitiveFieldFilterProvider() logging.SensitiveFieldFilter {
	return logging.NewDefaultSensitiveFieldFilter()
}

// Enhanced filter with custom patterns
func EnhancedSensitiveFieldFilterProvider() logging.SensitiveFieldFilter {
	return logging.NewEnhancedSensitiveFieldFilter()
}

// Wire provider sets for different configurations

// BasicProviderSet provides basic config and logging
var BasicProviderSet = wire.NewSet(
	ConfigProvider,
	LoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
)

// HotReloadProviderSet provides config with hot-reload capability
var HotReloadProviderSet = wire.NewSet(
	ConfigWithWatcherProvider,
	LoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
)

// DevelopmentProviderSet provides development-optimized dependencies
var DevelopmentProviderSet = wire.NewSet(
	ConfigProvider,
	DevelopmentLoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
)

// ProductionProviderSet provides production-optimized dependencies with enhanced security
var ProductionProviderSet = wire.NewSet(
	ConfigProvider,
	ProductionLoggerProvider,
	ContextLoggerProvider,
	EnhancedSensitiveFieldFilterProvider,
)

// AllProviderSet includes all providers for maximum flexibility
var AllProviderSet = wire.NewSet(
	ConfigProvider,
	ConfigWithWatcherProvider,
	LoggerProvider,
	DevelopmentLoggerProvider,
	ProductionLoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
	EnhancedSensitiveFieldFilterProvider,
)

// Application dependency structure for Wire
type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	ContextLogger logging.ContextLogger
	Filter        logging.SensitiveFieldFilter
}

// ApplicationWithWatcher includes hot-reload capability
type ApplicationWithWatcher struct {
	ConfigWatcher *config.ConfigWithWatcher
	Logger        *zap.Logger
	ContextLogger logging.ContextLogger
	Filter        logging.SensitiveFieldFilter
}

// NewApplication creates a new application with basic dependencies
func NewApplication(
	cfg *config.Config,
	logger *zap.Logger,
	contextLogger logging.ContextLogger,
	filter logging.SensitiveFieldFilter,
) *Application {
	return &Application{
		Config:        cfg,
		Logger:        logger,
		ContextLogger: contextLogger,
		Filter:        filter,
	}
}

// NewApplicationWithWatcher creates a new application with hot-reload capability
func NewApplicationWithWatcher(
	configWatcher *config.ConfigWithWatcher,
	logger *zap.Logger,
	contextLogger logging.ContextLogger,
	filter logging.SensitiveFieldFilter,
) *ApplicationWithWatcher {
	return &ApplicationWithWatcher{
		ConfigWatcher: configWatcher,
		Logger:        logger,
		ContextLogger: contextLogger,
		Filter:        filter,
	}
}

// Wire sets for application creation
var ApplicationProviderSet = wire.NewSet(
	BasicProviderSet,
	NewApplication,
)

var ApplicationWithWatcherProviderSet = wire.NewSet(
	HotReloadProviderSet,
	wire.Bind(new(*config.Config), new(*config.ConfigWithWatcher)),
	NewApplicationWithWatcher,
)

// Provider functions for specific use cases

// WebServerProviderSet provides dependencies optimized for web server use
var WebServerProviderSet = wire.NewSet(
	ConfigProvider,
	ProductionLoggerProvider,
	ContextLoggerProvider,
	EnhancedSensitiveFieldFilterProvider,
)

// CLIProviderSet provides dependencies optimized for CLI applications
var CLIProviderSet = wire.NewSet(
	ConfigProvider,
	DevelopmentLoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
)

// TestingProviderSet provides dependencies optimized for testing
var TestingProviderSet = wire.NewSet(
	ConfigProvider,
	DevelopmentLoggerProvider,
	ContextLoggerProvider,
	SensitiveFieldFilterProvider,
)

// Close gracefully shuts down the application dependencies
func (app *Application) Close() error {
	if app.Logger != nil {
		return app.Logger.Sync()
	}
	return nil
}

// Close gracefully shuts down the application with watcher
func (app *ApplicationWithWatcher) Close() error {
	var err error

	// Close config watcher
	if app.ConfigWatcher != nil {
		if closeErr := app.ConfigWatcher.Close(); closeErr != nil {
			err = closeErr
		}
	}

	// Sync logger
	if app.Logger != nil {
		if syncErr := app.Logger.Sync(); syncErr != nil && err == nil {
			err = syncErr
		}
	}

	return err
}

// GetConfig returns the current configuration
func (app *ApplicationWithWatcher) GetConfig() *config.Config {
	if app.ConfigWatcher != nil {
		return app.ConfigWatcher.GetConfig()
	}
	return nil
}
