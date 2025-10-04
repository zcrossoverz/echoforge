package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMissingConfigFile tests behavior when config.yaml doesn't exist
func TestMissingConfigFile(t *testing.T) {
	// Create a temporary directory without config file
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Should use defaults and environment variables
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	config, err := NewConfig()
	require.NoError(t, err, "Should create config with defaults when file missing")

	assert.Equal(t, "postgres://user:pass@localhost:5432/testdb", config.DBDSN)
	assert.Equal(t, "test-secret-key-at-least-32-characters", config.JWTSecret)
	assert.Equal(t, "info", config.LogLevel) // Default
}

// TestMalformedYAML tests behavior with invalid YAML syntax
func TestMalformedYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write malformed YAML
	malformedYAML := `
db_dsn: "postgres://user:pass@localhost:5432/testdb"
jwt_secret: "test-secret-key-at-least-32-characters"
log_level: info
invalid_yaml: [unclosed bracket
	another_field: value
`

	err := os.WriteFile(configPath, []byte(malformedYAML), 0644)
	require.NoError(t, err)

	// Change to temp directory so config.yaml is found
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	_, err = NewConfig()
	assert.Error(t, err, "Should fail with malformed YAML")
	assert.Contains(t, err.Error(), "parse", "Error should mention parsing issue")
}

// TestInvalidEnvironmentValues tests various invalid environment variable values
func TestInvalidEnvironmentValues(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty_db_dsn",
			envVars: map[string]string{
				"DB_DSN":     "",
				"JWT_SECRET": "test-secret-key-at-least-32-characters",
			},
			wantErr: true,
			errMsg:  "DBDSN is required",
		},
		{
			name: "invalid_db_dsn_format",
			envVars: map[string]string{
				"DB_DSN":     "not-a-valid-url",
				"JWT_SECRET": "test-secret-key-at-least-32-characters",
			},
			wantErr: true,
			errMsg:  "DBDSN must be a valid URL",
		},
		{
			name: "short_jwt_secret",
			envVars: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/testdb",
				"JWT_SECRET": "short",
			},
			wantErr: true,
			errMsg:  "JWTSecret must be at least 32 characters long",
		},
		{
			name: "invalid_log_level",
			envVars: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/testdb",
				"JWT_SECRET": "test-secret-key-at-least-32-characters",
				"LOG_LEVEL":  "invalid",
			},
			wantErr: true,
			errMsg:  "LogLevel must be one of: debug info error",
		},
		{
			name: "invalid_enable_hot_reload",
			envVars: map[string]string{
				"DB_DSN":            "postgres://user:pass@localhost:5432/testdb",
				"JWT_SECRET":        "test-secret-key-at-least-32-characters",
				"ENABLE_HOT_RELOAD": "maybe",
			},
			wantErr: false, // Should use default false for invalid boolean
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := NewConfig()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

// TestCorruptedConfigFile tests behavior with corrupted or unreadable config files
func TestCorruptedConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a file with valid YAML but invalid content
	corruptedYAML := `
db_dsn: 123  # Should be string
jwt_secret: []  # Should be string
log_level: true  # Should be string
enable_hot_reload: "not_a_boolean"  # Should be boolean
`

	err := os.WriteFile(configPath, []byte(corruptedYAML), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	_, err = NewConfig()
	assert.Error(t, err, "Should fail with type mismatch")
}

// TestPermissionDeniedConfigFile tests behavior when config file exists but can't be read
func TestPermissionDeniedConfigFile(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config file
	validYAML := `
db_dsn: "postgres://user:pass@localhost:5432/testdb"
jwt_secret: "test-secret-key-at-least-32-characters"
log_level: "info"
`

	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	require.NoError(t, err)

	// Remove read permissions
	err = os.Chmod(configPath, 0000)
	require.NoError(t, err)

	// Restore permissions for cleanup
	defer os.Chmod(configPath, 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	_, err = NewConfig()
	assert.Error(t, err, "Should fail when config file can't be read")
}

// TestExtremelyLargeConfigFile tests behavior with very large config files
func TestExtremelyLargeConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a large YAML file (but still valid)
	largeYAML := `
db_dsn: "postgres://user:pass@localhost:5432/testdb"
jwt_secret: "test-secret-key-at-least-32-characters"
log_level: "info"
`

	// Add many extra fields to make it large
	for i := 0; i < 1000; i++ {
		largeYAML += "extra_field_" + string(rune(i)) + ": \"value\"\n"
	}

	err := os.WriteFile(configPath, []byte(largeYAML), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	start := time.Now()
	config, err := NewConfig()
	duration := time.Since(start)

	// Should handle large files gracefully (within 5 seconds per requirement)
	assert.NoError(t, err, "Should handle large config files")
	assert.NotNil(t, config)
	assert.Less(t, duration, 5*time.Second, "Config loading should be under 5 seconds")
}

// TestConcurrentConfigCreation tests concurrent access to config creation
func TestConcurrentConfigCreation(t *testing.T) {
	// Set valid environment variables
	os.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
	}()

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Launch multiple goroutines trying to create config simultaneously
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := NewConfig()
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent config creation should not fail")
	}
}

// TestEnvironmentVariablePrecedence tests that environment variables override config file
func TestEnvironmentVariablePrecedence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config file with one set of values
	configYAML := `
db_dsn: "postgres://file:pass@localhost:5432/filedb"
jwt_secret: "file-secret-key-at-least-32-characters"
log_level: "debug"
enable_hot_reload: true
`

	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	// Set different values in environment
	os.Setenv("DB_DSN", "postgres://env:pass@localhost:5432/envdb")
	os.Setenv("JWT_SECRET", "env-secret-key-at-least-32-characters")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config, err := NewConfig()
	require.NoError(t, err)

	// Environment variables should override file values
	assert.Equal(t, "postgres://env:pass@localhost:5432/envdb", config.DBDSN)
	assert.Equal(t, "env-secret-key-at-least-32-characters", config.JWTSecret)
	assert.Equal(t, "error", config.LogLevel)
	assert.False(t, config.EnableHotReload)
}
