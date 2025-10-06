package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    map[string]string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name: "valid config from environment variables",
			setupEnv: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/test?sslmode=disable",
				"JWT_SECRET": "super-secret-jwt-key-at-least-32-characters-long",
				"LOG_LEVEL":  "debug",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://user:pass@localhost:5432/test?sslmode=disable", cfg.DBDSN)
				assert.Equal(t, "super-secret-jwt-key-at-least-32-characters-long", cfg.JWTSecret)
				assert.Equal(t, "debug", cfg.LogLevel)
			},
		},
		{
			name: "valid config with defaults",
			setupEnv: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/test?sslmode=disable",
				"JWT_SECRET": "super-secret-jwt-key-at-least-32-characters-long",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "info", cfg.LogLevel) // Should default to "info"
				assert.False(t, cfg.EnableHotReload)  // Should default to false
			},
		},
		{
			name: "missing required DB_DSN",
			setupEnv: map[string]string{
				"JWT_SECRET": "super-secret-jwt-key-at-least-32-characters-long",
			},
			expectError: true,
		},
		{
			name: "missing required JWT_SECRET",
			setupEnv: map[string]string{
				"DB_DSN": "postgres://user:pass@localhost:5432/test?sslmode=disable",
			},
			expectError: true,
		},
		{
			name: "JWT secret too short",
			setupEnv: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/test?sslmode=disable",
				"JWT_SECRET": "short",
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			setupEnv: map[string]string{
				"DB_DSN":     "postgres://user:pass@localhost:5432/test?sslmode=disable",
				"JWT_SECRET": "super-secret-jwt-key-at-least-32-characters-long",
				"LOG_LEVEL":  "verbose",
			},
			expectError: true,
		},
		{
			name: "invalid DB_DSN format",
			setupEnv: map[string]string{
				"DB_DSN":     "not-a-valid-url",
				"JWT_SECRET": "super-secret-jwt-key-at-least-32-characters-long",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Clearenv()

			// Set up environment variables
			for key, value := range tt.setupEnv {
				os.Setenv(key, value)
			}

			// Call the function under test
			cfg, err := NewConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestNewConfigYAMLLoading(t *testing.T) {
	// This test will verify YAML loading with various configurations
	// It should test precedence: env > YAML > defaults
	t.Skip("NewConfig() factory not implemented yet - test should fail")
}

func TestNewConfigHotReload(t *testing.T) {
	// This test will verify hot-reload functionality
	t.Skip("Config hot-reload not implemented yet - test should fail")
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				DBDSN:     "postgres://user:pass@localhost:5432/db?sslmode=disable",
				JWTSecret: "super-secret-jwt-key-at-least-32-characters-long",
				LogLevel:  "info",
			},
			valid: true,
		},
		{
			name: "empty DB_DSN",
			config: Config{
				JWTSecret: "super-secret-jwt-key-at-least-32-characters-long",
				LogLevel:  "info",
			},
			valid: false,
		},
		{
			name: "short JWT secret",
			config: Config{
				DBDSN:     "postgres://user:pass@localhost:5432/db?sslmode=disable",
				JWTSecret: "short",
				LogLevel:  "info",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will test the validation function when implemented
			t.Skip("Config validation not implemented yet - test should fail")
		})
	}
}
