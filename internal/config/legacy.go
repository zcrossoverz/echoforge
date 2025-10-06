package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// LegacyConfig provides backward compatibility for existing config access patterns
type LegacyConfig struct {
	config *Config
	mu     sync.RWMutex
}

// Global legacy instance for backward compatibility
var legacyInstance *LegacyConfig
var legacyOnce sync.Once

// InitLegacyConfig initializes the legacy config with the new config system
func InitLegacyConfig(cfg *Config) {
	legacyOnce.Do(func() {
		legacyInstance = &LegacyConfig{
			config: cfg,
		}
	})
}

// GetLegacyConfig returns the global legacy config instance
func GetLegacyConfig() *LegacyConfig {
	if legacyInstance == nil {
		// Auto-initialize with a new config if not already done
		cfg, err := NewConfig()
		if err != nil {
			// Fallback to environment-only config
			cfg = &Config{
				DBDSN:           os.Getenv("DB_DSN"),
				JWTSecret:       os.Getenv("JWT_SECRET"),
				LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
				EnableHotReload: getBoolEnvOrDefault("ENABLE_HOT_RELOAD", false),
			}
		}
		InitLegacyConfig(cfg)
	}
	return legacyInstance
}

// UpdateConfig updates the underlying config (for hot-reload compatibility)
func (l *LegacyConfig) UpdateConfig(cfg *Config) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config = cfg
}

// Legacy getter methods for backward compatibility

// GetDBDSN returns the database DSN
func (l *LegacyConfig) GetDBDSN() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.DBDSN
}

// GetJWTSecret returns the JWT secret
func (l *LegacyConfig) GetJWTSecret() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.JWTSecret
}

// GetLogLevel returns the log level
func (l *LegacyConfig) GetLogLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.LogLevel
}

// IsHotReloadEnabled returns whether hot-reload is enabled
func (l *LegacyConfig) IsHotReloadEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.EnableHotReload
}

// GetConfig returns the underlying config struct
func (l *LegacyConfig) GetConfig() *Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	// Return a copy to prevent external modifications
	configCopy := *l.config
	return &configCopy
}

// Global legacy functions for existing code patterns

// GetDBDSN legacy global function
func GetDBDSN() string {
	return GetLegacyConfig().GetDBDSN()
}

// GetJWTSecret legacy global function
func GetJWTSecret() string {
	return GetLegacyConfig().GetJWTSecret()
}

// GetLogLevel legacy global function
func GetLogLevel() string {
	return GetLegacyConfig().GetLogLevel()
}

// IsHotReloadEnabled legacy global function
func IsHotReloadEnabled() bool {
	return GetLegacyConfig().IsHotReloadEnabled()
}

// Legacy environment variable patterns

// GetEnv returns an environment variable with a fallback
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetIntEnv returns an environment variable as integer with fallback
func GetIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetBoolEnv returns an environment variable as boolean with fallback
func GetBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return fallback
}

// Legacy config struct for existing code that expects specific structure
type LegacyConfigStruct struct {
	Database struct {
		DSN string `json:"dsn"`
	} `json:"database"`
	JWT struct {
		Secret string `json:"secret"`
	} `json:"jwt"`
	Log struct {
		Level string `json:"level"`
	} `json:"log"`
	Features struct {
		HotReload bool `json:"hot_reload"`
	} `json:"features"`
}

// ToLegacyStruct converts the new config to legacy structure
func (l *LegacyConfig) ToLegacyStruct() *LegacyConfigStruct {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &LegacyConfigStruct{
		Database: struct {
			DSN string `json:"dsn"`
		}{
			DSN: l.config.DBDSN,
		},
		JWT: struct {
			Secret string `json:"secret"`
		}{
			Secret: l.config.JWTSecret,
		},
		Log: struct {
			Level string `json:"level"`
		}{
			Level: l.config.LogLevel,
		},
		Features: struct {
			HotReload bool `json:"hot_reload"`
		}{
			HotReload: l.config.EnableHotReload,
		},
	}
}

// GetLegacyStruct returns the legacy config structure
func GetLegacyStruct() *LegacyConfigStruct {
	return GetLegacyConfig().ToLegacyStruct()
}

// Legacy initialization patterns

// MustLoadConfig loads config and panics on error (legacy pattern)
func MustLoadConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	InitLegacyConfig(cfg)
	return cfg
}

// LoadConfigOrPanic legacy alias
func LoadConfigOrPanic() *Config {
	return MustLoadConfig()
}

// Configuration migration helpers

// ConfigMigration helps migrate from old config patterns to new ones
type ConfigMigration struct {
	warnings []string
}

// NewConfigMigration creates a new config migration helper
func NewConfigMigration() *ConfigMigration {
	return &ConfigMigration{
		warnings: make([]string, 0),
	}
}

// CheckLegacyEnvVars checks for deprecated environment variables
func (m *ConfigMigration) CheckLegacyEnvVars() []string {
	legacyVars := map[string]string{
		"DATABASE_URL":  "DB_DSN",
		"JWT_TOKEN":     "JWT_SECRET",
		"LOGGING_LEVEL": "LOG_LEVEL",
		"HOT_RELOAD":    "ENABLE_HOT_RELOAD",
		"DEBUG_MODE":    "LOG_LEVEL (use 'debug')",
	}

	warnings := make([]string, 0)
	for oldVar, newVar := range legacyVars {
		if os.Getenv(oldVar) != "" {
			warning := fmt.Sprintf("Deprecated environment variable '%s' found. Please use '%s' instead.", oldVar, newVar)
			warnings = append(warnings, warning)
		}
	}

	return warnings
}

// ApplyLegacyMigrations applies legacy environment variable migrations
func (m *ConfigMigration) ApplyLegacyMigrations() {
	// Migrate DATABASE_URL to DB_DSN
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" && os.Getenv("DB_DSN") == "" {
		os.Setenv("DB_DSN", dbURL)
	}

	// Migrate JWT_TOKEN to JWT_SECRET
	if jwtToken := os.Getenv("JWT_TOKEN"); jwtToken != "" && os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", jwtToken)
	}

	// Migrate LOGGING_LEVEL to LOG_LEVEL
	if logLevel := os.Getenv("LOGGING_LEVEL"); logLevel != "" && os.Getenv("LOG_LEVEL") == "" {
		os.Setenv("LOG_LEVEL", logLevel)
	}

	// Migrate HOT_RELOAD to ENABLE_HOT_RELOAD
	if hotReload := os.Getenv("HOT_RELOAD"); hotReload != "" && os.Getenv("ENABLE_HOT_RELOAD") == "" {
		os.Setenv("ENABLE_HOT_RELOAD", hotReload)
	}

	// Migrate DEBUG_MODE to LOG_LEVEL
	if debugMode := os.Getenv("DEBUG_MODE"); debugMode != "" && os.Getenv("LOG_LEVEL") == "" {
		if strings.ToLower(debugMode) == "true" || debugMode == "1" {
			os.Setenv("LOG_LEVEL", "debug")
		}
	}
}

// Helper functions for legacy compatibility

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

// Legacy callback system for config changes
type LegacyConfigCallback func(oldConfig, newConfig *Config)

var legacyCallbacks []LegacyConfigCallback
var callbackMu sync.RWMutex

// RegisterLegacyCallback registers a callback for config changes
func RegisterLegacyCallback(callback LegacyConfigCallback) {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	legacyCallbacks = append(legacyCallbacks, callback)
}

// TriggerLegacyCallbacks triggers all registered callbacks
func TriggerLegacyCallbacks(oldConfig, newConfig *Config) {
	callbackMu.RLock()
	callbacks := make([]LegacyConfigCallback, len(legacyCallbacks))
	copy(callbacks, legacyCallbacks)
	callbackMu.RUnlock()

	for _, callback := range callbacks {
		callback(oldConfig, newConfig)
	}
}

// SetupLegacyCompatibility sets up full backward compatibility
func SetupLegacyCompatibility(cfg *Config) {
	// Initialize legacy config
	InitLegacyConfig(cfg)

	// Apply legacy migrations
	migration := NewConfigMigration()
	migration.ApplyLegacyMigrations()

	// Check for warnings
	if warnings := migration.CheckLegacyEnvVars(); len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration warnings:\n")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", warning)
		}
	}
}

// Legacy hot-reload integration
func SetupLegacyHotReload(watcher *ConfigWatcher) error {
	if watcher == nil {
		return nil // Hot-reload not enabled
	}

	// Register callback to update legacy config
	watcher.AddReloadCallback(func(oldConfig, newConfig *Config) error {
		// Update legacy instance
		if legacyInstance != nil {
			legacyInstance.UpdateConfig(newConfig)
		}

		// Trigger legacy callbacks
		TriggerLegacyCallbacks(oldConfig, newConfig)

		return nil
	})

	return nil
}
