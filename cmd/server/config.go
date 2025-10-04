package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zcrossoverz/echoforge/internal/config"
)

// ValidateServerConfig validates server configuration
func ValidateServerConfig(cfg *config.Config) error {
	var validationErrors []string

	// Validate database configuration
	if cfg.DBDSN == "" {
		validationErrors = append(validationErrors, "database DSN is required")
	} else {
		if err := validateDatabaseDSN(cfg.DBDSN); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid database DSN: %v", err))
		}
	}

	// Validate JWT configuration
	if cfg.JWTSecret == "" {
		validationErrors = append(validationErrors, "JWT secret is required")
	} else if len(cfg.JWTSecret) < 32 {
		validationErrors = append(validationErrors, "JWT secret must be at least 32 characters long")
	}

	// JWT configuration is already validated by the config package
	// No additional validation needed here for JWT expiration

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}
	if !contains(validLogLevels, strings.ToLower(cfg.LogLevel)) {
		validationErrors = append(validationErrors, fmt.Sprintf("invalid log level '%s', must be one of: %s", cfg.LogLevel, strings.Join(validLogLevels, ", ")))
	}

	// Site ID validation - check environment variable if set
	if siteID := os.Getenv("SITE_ID"); siteID != "" {
		if err := validateSiteID(siteID); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid site ID: %v", err))
		}
	}

	// Validate environment-specific settings
	if err := validateEnvironmentSettings(); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("environment validation failed: %v", err))
	}

	// Return validation errors if any
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(validationErrors, "\n- "))
	}

	return nil
}

// validateDatabaseDSN validates the database connection string
func validateDatabaseDSN(dsn string) error {
	if !strings.Contains(dsn, "postgresql://") && !strings.Contains(dsn, "postgres://") {
		return errors.New("DSN must be a PostgreSQL connection string")
	}

	// Basic validation for required components
	if !strings.Contains(dsn, "@") {
		return errors.New("DSN must contain authentication information")
	}

	if !strings.Contains(dsn, "/") {
		return errors.New("DSN must contain database name")
	}

	return nil
}

// validateSiteID validates the site identifier
func validateSiteID(siteID string) error {
	if len(siteID) < 3 {
		return errors.New("site ID must be at least 3 characters long")
	}

	if len(siteID) > 50 {
		return errors.New("site ID must not exceed 50 characters")
	}

	// Site ID should contain only alphanumeric characters and hyphens
	for _, char := range siteID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return errors.New("site ID can only contain alphanumeric characters, hyphens, and underscores")
		}
	}

	return nil
}

// validateEnvironmentSettings validates environment-specific configuration
func validateEnvironmentSettings() error {
	// Validate PORT if set
	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid PORT environment variable: %v", err)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("PORT must be between 1 and 65535, got %d", port)
		}
	}

	// Validate ENV if set
	if env := os.Getenv("ENV"); env != "" {
		validEnvs := []string{"development", "staging", "production", "test"}
		if !contains(validEnvs, strings.ToLower(env)) {
			return fmt.Errorf("invalid ENV '%s', must be one of: %s", env, strings.Join(validEnvs, ", "))
		}
	}

	// Validate LOG_LEVEL if set
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}
		if !contains(validLevels, strings.ToLower(logLevel)) {
			return fmt.Errorf("invalid LOG_LEVEL '%s', must be one of: %s", logLevel, strings.Join(validLevels, ", "))
		}
	}

	return nil
}

// ValidateRequiredEnvironmentVariables ensures critical environment variables are set for production
func ValidateRequiredEnvironmentVariables() error {
	env := strings.ToLower(os.Getenv("ENV"))

	// For production, certain environment variables are mandatory
	if env == "production" {
		requiredVars := []string{
			"DATABASE_URL",
			"JWT_SECRET",
		}

		var missingVars []string
		for _, varName := range requiredVars {
			if os.Getenv(varName) == "" {
				missingVars = append(missingVars, varName)
			}
		}

		if len(missingVars) > 0 {
			return fmt.Errorf("required environment variables missing in production: %s", strings.Join(missingVars, ", "))
		}
	}

	return nil
}

// ValidateSystemRequirements checks system-level requirements
func ValidateSystemRequirements() error {
	// Check if required directories exist and are writable
	requiredDirs := []string{
		"./logs", // For log files (if file logging is enabled)
		"./temp", // For temporary files
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create required directory '%s': %v", dir, err)
			}
		}
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
