package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config represents the application configuration with validation rules
type Config struct {
	// Database Configuration (required)
	DBDSN string `mapstructure:"DB_DSN" validate:"required,url"`

	// Authentication Configuration (required, minimum 32 characters)
	JWTSecret string `mapstructure:"JWT_SECRET" validate:"required,min=32"`

	// Logging Configuration (optional, defaults to "info")
	LogLevel string `mapstructure:"LOG_LEVEL" validate:"oneof=debug info error"`

	// Development Features (optional, defaults to false)
	EnableHotReload bool `mapstructure:"ENABLE_HOT_RELOAD"`
}

// validator instance for configuration validation
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// NewConfig creates and validates a new configuration instance
// Loads from environment variables (highest priority) -> config.yaml (medium) -> defaults (lowest)
func NewConfig() (*Config, error) {
	// Initialize a new Viper instance for isolation
	v := viper.New()

	// Set configuration file settings
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Enable automatic environment variable binding
	v.AutomaticEnv()

	// Explicitly bind environment variables to config keys
	v.BindEnv("DB_DSN")
	v.BindEnv("JWT_SECRET")
	v.BindEnv("LOG_LEVEL")
	v.BindEnv("ENABLE_HOT_RELOAD")

	// Set default values
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("ENABLE_HOT_RELOAD", false)

	// Read configuration file (if it exists)
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error - we can rely on env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal configuration into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates the configuration struct using go-playground/validator
func validateConfig(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		// Provide more user-friendly error messages
		var validationErrors []string

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", err.Field()))
			case "min":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param()))
			case "url":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid URL", err.Field()))
			case "oneof":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
			default:
				validationErrors = append(validationErrors, fmt.Sprintf("%s validation failed: %s", err.Field(), err.Tag()))
			}
		}

		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, ", "))
	}

	return nil
}

// GetConfigPath returns the path to the configuration file being used
func GetConfigPath() string {
	// Check common config file locations
	configPaths := []string{
		"./configs/config.yaml",
		"./config.yaml",
		"./configs/config.yml",
		"./config.yml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// ConfigLoader defines the interface for configuration management
type ConfigLoader interface {
	// Load configuration from all sources with validation
	Load() (*Config, error)

	// EnableHotReload starts watching config file for changes (development only)
	EnableHotReload(callback func(*Config)) error

	// Validate checks configuration against business rules
	Validate(config *Config) error
}

// DefaultConfigLoader implements ConfigLoader interface
type DefaultConfigLoader struct{}

// Load implements ConfigLoader.Load
func (d *DefaultConfigLoader) Load() (*Config, error) {
	return NewConfig()
}

// EnableHotReload implements ConfigLoader.EnableHotReload
func (d *DefaultConfigLoader) EnableHotReload(callback func(*Config)) error {
	// Hot-reload functionality will be implemented in watcher.go
	return fmt.Errorf("hot-reload functionality not yet implemented")
}

// Validate implements ConfigLoader.Validate
func (d *DefaultConfigLoader) Validate(config *Config) error {
	return validateConfig(config)
}
