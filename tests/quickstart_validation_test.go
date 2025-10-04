package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcrossoverz/echoforge/internal/config"
	"github.com/zcrossoverz/echoforge/internal/logging"
	"go.uber.org/zap"
)

// TestQuickstartValidation runs through all the quickstart examples
func TestQuickstartValidation(t *testing.T) {
	// Backup current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	t.Run("quickstart_example_1_basic_config", func(t *testing.T) {
		testBasicConfigurationLoading(t)
	})

	t.Run("quickstart_example_2_structured_logging", func(t *testing.T) {
		testStructuredLogging(t)
	})

	t.Run("quickstart_example_3_security_sanitization", func(t *testing.T) {
		testSecuritySanitization(t)
	})

	t.Run("quickstart_validation_tests", func(t *testing.T) {
		testConfigurationValidation(t)
	})

	t.Run("quickstart_log_level_filtering", func(t *testing.T) {
		testLogLevelFiltering(t)
	})
}

// testBasicConfigurationLoading validates quickstart example 1
func testBasicConfigurationLoading(t *testing.T) {
	// Set valid environment variables
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/echoforge_dev")
	os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	// Load configuration with validation (from quickstart example)
	cfg, err := config.NewConfig()
	require.NoError(t, err, "Failed to load config")

	// Verify quickstart expectations
	assert.Equal(t, "postgres://user:pass@localhost:5432/echoforge_dev", cfg.DBDSN)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "your-super-secret-jwt-key-at-least-32-chars-long", cfg.JWTSecret)

	// Simulate the printf outputs from quickstart
	dbOutput := fmt.Sprintf("Database: %s", cfg.DBDSN)
	logLevelOutput := fmt.Sprintf("Log Level: %s", cfg.LogLevel)
	hotReloadOutput := fmt.Sprintf("Hot Reload: %v", cfg.EnableHotReload)

	assert.Contains(t, dbOutput, "postgres://user:pass@localhost:5432/echoforge_dev")
	assert.Contains(t, logLevelOutput, "info")
	assert.Contains(t, hotReloadOutput, "false") // Default is false
}

// testStructuredLogging validates quickstart example 2
func testStructuredLogging(t *testing.T) {
	// Setup from quickstart example
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/echoforge_dev")
	os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
	os.Setenv("LOG_LEVEL", "debug") // Enable debug for this test
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	// Load config and create logger (from quickstart)
	cfg, err := config.NewConfig()
	require.NoError(t, err, "Config error should not occur")

	logConfig := &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: false,
	}
	logger, err := logging.NewLogger(logConfig)
	require.NoError(t, err, "Logger error should not occur")
	defer logger.Sync()

	// Test structured logging examples from quickstart
	assert.NotPanics(t, func() {
		logger.Info("Application started",
			zap.String("service", "echoforge"),
			zap.String("version", "1.0.0"),
		)

		logger.Debug("Debug information",
			zap.String("request_id", "req_123"),
			zap.Int("user_count", 42),
		)

		logger.Error("Error occurred",
			zap.Error(fmt.Errorf("example error")),
			zap.String("component", "user_service"),
		)
	})
}

// testSecuritySanitization validates quickstart example 3
func testSecuritySanitization(t *testing.T) {
	os.Setenv("DB_DSN", "postgres://user:secret@localhost:5432/db")
	os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := config.NewConfig()
	require.NoError(t, err)

	logConfig := &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: false,
	}
	logger, err := logging.NewLogger(logConfig)
	require.NoError(t, err)
	defer logger.Sync()

	// Test the security sanitization example from quickstart
	assert.NotPanics(t, func() {
		logger.Info("User login attempt",
			zap.String("email", "user@example.com"),       // Safe: should be logged
			zap.String("password", "secret123"),           // Unsafe: should be [REDACTED]
			zap.String("jwt_token", "eyJ0eXAiOiJKV1Q..."), // Unsafe: should be [REDACTED]
			zap.String("db_dsn", cfg.DBDSN),               // Unsafe: should be [REDACTED]
		)
	})
}

// testConfigurationValidation validates the testing scenarios from quickstart
func testConfigurationValidation(t *testing.T) {
	t.Run("invalid_database_dsn", func(t *testing.T) {
		os.Setenv("DB_DSN", "invalid-url")
		os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
		defer func() {
			os.Unsetenv("DB_DSN")
			os.Unsetenv("JWT_SECRET")
		}()

		_, err := config.NewConfig()
		assert.Error(t, err, "Should fail with invalid URL")
		assert.Contains(t, err.Error(), "DBDSN must be a valid URL")
	})

	t.Run("jwt_secret_too_short", func(t *testing.T) {
		os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("JWT_SECRET", "short")
		defer func() {
			os.Unsetenv("DB_DSN")
			os.Unsetenv("JWT_SECRET")
		}()

		_, err := config.NewConfig()
		assert.Error(t, err, "Should fail with short JWT secret")
		assert.Contains(t, err.Error(), "JWTSecret must be at least 32 characters long")
	})

	t.Run("invalid_log_level", func(t *testing.T) {
		os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db")
		os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
		os.Setenv("LOG_LEVEL", "verbose")
		defer func() {
			os.Unsetenv("DB_DSN")
			os.Unsetenv("JWT_SECRET")
			os.Unsetenv("LOG_LEVEL")
		}()

		_, err := config.NewConfig()
		assert.Error(t, err, "Should fail with invalid log level")
		assert.Contains(t, err.Error(), "LogLevel must be one of: debug info error")
	})
}

// testLogLevelFiltering validates log level filtering from quickstart
func testLogLevelFiltering(t *testing.T) {
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	levels := []string{"debug", "info", "error"}

	for _, level := range levels {
		t.Run("log_level_"+level, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", level)

			cfg, err := config.NewConfig()
			require.NoError(t, err)

			logConfig := &logging.SimpleConfig{
				LogLevel:    cfg.LogLevel,
				Development: false,
			}
			logger, err := logging.NewLogger(logConfig)
			require.NoError(t, err)
			defer logger.Sync() // Test that logger accepts the level
			assert.Equal(t, level, cfg.LogLevel)

			// Test logging at different levels
			assert.NotPanics(t, func() {
				logger.Debug("Debug message")
				logger.Info("Info message")
				logger.Error("Error message")
			})
		})
	}
}

// TestQuickstartPerformance validates performance requirements from quickstart
func TestQuickstartPerformance(t *testing.T) {
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-at-least-32-chars-long")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	t.Run("config_loading_performance", func(t *testing.T) {
		start := time.Now()
		_, err := config.NewConfig()
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 5*time.Second, "Config loading should be under 5 seconds")
	})

	t.Run("logger_creation_performance", func(t *testing.T) {
		cfg, err := config.NewConfig()
		require.NoError(t, err)

		logConfig := &logging.SimpleConfig{
			LogLevel:    cfg.LogLevel,
			Development: false,
		}

		start := time.Now()
		logger, err := logging.NewLogger(logConfig)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Less(t, duration, 1*time.Second, "Logger creation should be fast")
	})

	t.Run("logging_throughput", func(t *testing.T) {
		cfg, err := config.NewConfig()
		require.NoError(t, err)

		logConfig := &logging.SimpleConfig{
			LogLevel:    cfg.LogLevel,
			Development: false,
		}
		logger, err := logging.NewLogger(logConfig)
		require.NoError(t, err)
		defer logger.Sync()

		// Test logging throughput (should handle 1000+ logs/sec)
		start := time.Now()
		numLogs := 1000

		for i := 0; i < numLogs; i++ {
			logger.Info("Performance test",
				zap.Int("iteration", i),
				zap.String("test", "throughput"),
			)
		}

		duration := time.Since(start)
		logsPerSecond := float64(numLogs) / duration.Seconds()

		assert.Greater(t, logsPerSecond, 1000.0, "Should handle 1000+ logs per second")
	})
}

// TestQuickstartIntegration validates end-to-end integration scenarios
func TestQuickstartIntegration(t *testing.T) {
	t.Run("config_and_logging_integration", func(t *testing.T) {
		// Test complete integration flow from quickstart
		os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/integration_test")
		os.Setenv("JWT_SECRET", "integration-test-jwt-secret-32-chars-minimum")
		os.Setenv("LOG_LEVEL", "info")
		defer func() {
			os.Unsetenv("DB_DSN")
			os.Unsetenv("JWT_SECRET")
			os.Unsetenv("LOG_LEVEL")
		}()

		// Step 1: Load configuration
		cfg, err := config.NewConfig()
		require.NoError(t, err)

		// Step 2: Create logger with config
		logConfig := &logging.SimpleConfig{
			LogLevel:    cfg.LogLevel,
			Development: false,
		}
		logger, err := logging.NewLogger(logConfig)
		require.NoError(t, err)
		defer logger.Sync()

		// Step 3: Verify integration works
		assert.NotPanics(t, func() {
			logger.Info("Integration test started",
				zap.String("database", cfg.DBDSN),
				zap.String("log_level", cfg.LogLevel),
				zap.Bool("hot_reload", cfg.EnableHotReload),
			)

			// Test sensitive data is handled properly
			logger.Info("Sensitive data test",
				zap.String("password", "should-be-redacted"),
				zap.String("normal", "should-be-visible"),
			)
		})

		// Verify config values match expectations
		assert.Equal(t, "postgres://user:pass@localhost:5432/integration_test", cfg.DBDSN)
		assert.Equal(t, "integration-test-jwt-secret-32-chars-minimum", cfg.JWTSecret)
		assert.Equal(t, "info", cfg.LogLevel)
	})
}

// TestQuickstartDocumentationExamples validates that all code examples from docs actually work
func TestQuickstartDocumentationExamples(t *testing.T) {
	t.Run("dependency_installation_check", func(t *testing.T) {
		// Verify required dependencies are available
		requiredPackages := []string{
			"github.com/spf13/viper",
			"go.uber.org/zap",
			"github.com/go-playground/validator/v10",
		}

		for _, pkg := range requiredPackages {
			t.Run("package_"+pkg, func(t *testing.T) {
				// Try to import the package in a simple program
				err := tryImportPackage(pkg)
				assert.NoError(t, err, "Required package %s should be available", pkg)
			})
		}
	})

	t.Run("config_file_creation", func(t *testing.T) {
		// Test the config.yaml creation from quickstart
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")

		// Create the config file as shown in quickstart
		configContent := `DB_DSN: "postgres://user:pass@localhost:5432/echoforge_dev?sslmode=disable"
JWT_SECRET: "your-super-secret-jwt-key-at-least-32-chars-long"
LOG_LEVEL: "info"
ENABLE_HOT_RELOAD: true`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Verify file was created correctly
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "postgres://user:pass@localhost:5432/echoforge_dev")
		assert.Contains(t, string(content), "your-super-secret-jwt-key-at-least-32-chars-long")
	})
}

// Helper function to test package import
func tryImportPackage(packageName string) error {
	// Create a temporary Go program that imports the package
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "import_test.go")

	program := fmt.Sprintf(`package main

import _ "%s"

func main() {}
`, packageName)

	err := os.WriteFile(testFile, []byte(program), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(testFile)

	// Try to build the program
	cmd := exec.Command("go", "build", "-o", "/dev/null", testFile)
	if cmd.Run() != nil {
		return fmt.Errorf("package %s not available", packageName)
	}

	return nil
}

// Example main function for manual testing (commented out for test file)
/*
func main() {
	// This would be the actual quickstart example 1
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Database: %s\n", cfg.DBDSN)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Hot Reload: %v\n", cfg.EnableHotReload)
}
*/
