package logging

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestInvalidLogLevels tests error handling for invalid log levels
func TestInvalidLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty_log_level",
			logLevel: "",
			wantErr:  true,
			errMsg:   "invalid log level:",
		},
		{
			name:     "invalid_log_level",
			logLevel: "invalid",
			wantErr:  true,
			errMsg:   "invalid log level: invalid",
		},
		{
			name:     "case_sensitive_debug",
			logLevel: "DEBUG",
			wantErr:  true,
			errMsg:   "invalid log level: DEBUG",
		},
		{
			name:     "case_sensitive_info",
			logLevel: "INFO",
			wantErr:  true,
			errMsg:   "invalid log level: INFO",
		},
		{
			name:     "trace_level_not_supported",
			logLevel: "trace",
			wantErr:  true,
			errMsg:   "invalid log level: trace",
		},
		{
			name:     "warn_level_not_supported",
			logLevel: "warn",
			wantErr:  true,
			errMsg:   "invalid log level: warn",
		},
		{
			name:     "valid_debug",
			logLevel: "debug",
			wantErr:  false,
		},
		{
			name:     "valid_info",
			logLevel: "info",
			wantErr:  false,
		},
		{
			name:     "valid_error",
			logLevel: "error",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SimpleConfig{
				LogLevel:    tt.logLevel,
				Development: false,
			}

			logger, err := NewLogger(config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

// TestContextPropagationFailures tests error handling in context propagation
func TestContextPropagationFailures(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	_, err := NewLogger(config)
	require.NoError(t, err)

	t.Run("nil_context", func(t *testing.T) {
		// Should not panic with nil context
		assert.NotPanics(t, func() {
			WithRequestID(nil, "test-request-id")
		})
	})

	t.Run("empty_request_id", func(t *testing.T) {
		ctx := context.Background()

		// Should handle empty request ID gracefully
		assert.NotPanics(t, func() {
			newCtx := WithRequestID(ctx, "")
			// Should still create context even with empty ID
			assert.NotNil(t, newCtx)
		})
	})

	t.Run("very_long_request_id", func(t *testing.T) {
		ctx := context.Background()
		longID := strings.Repeat("a", 10000) // Very long request ID

		// Should handle very long request IDs without error
		assert.NotPanics(t, func() {
			newCtx := WithRequestID(ctx, longID)
			assert.NotNil(t, newCtx)

			// Should be able to retrieve it
			retrievedID := GetRequestID(newCtx)
			assert.Equal(t, longID, retrievedID)
		})
	})

	t.Run("special_characters_in_request_id", func(t *testing.T) {
		ctx := context.Background()
		specialID := "req-<>&\"'\\{}[]()!@#$%^&*"

		assert.NotPanics(t, func() {
			newCtx := WithRequestID(ctx, specialID)
			retrievedID := GetRequestID(newCtx)
			assert.Equal(t, specialID, retrievedID)
		})
	})

	t.Run("missing_request_id_from_context", func(t *testing.T) {
		ctx := context.Background()

		// Should return empty string for missing request ID, not panic
		assert.NotPanics(t, func() {
			requestID := GetRequestID(ctx)
			assert.Empty(t, requestID)
		})
	})
}

// TestSanitizationEdgeCases tests edge cases in sensitive data sanitization
func TestSanitizationEdgeCases(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	t.Run("nil_fields_slice", func(t *testing.T) {
		// Should not panic with nil fields
		assert.NotPanics(t, func() {
			logger.Info("Test message", nil...)
		})
	})

	t.Run("empty_field_values", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Info("Test message",
				zap.String("password", ""),
				zap.String("secret", ""),
				zap.String("token", ""),
			)
		})
	})

	t.Run("nested_sensitive_data", func(t *testing.T) {
		// Complex nested structures with sensitive data
		complexData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       123,
				"password": "secret123",
				"profile": map[string]interface{}{
					"email": "user@example.com",
					"token": "jwt-token-here",
				},
			},
		}

		assert.NotPanics(t, func() {
			logger.Info("Complex data", zap.Any("data", complexData))
		})
	})

	t.Run("very_large_field_values", func(t *testing.T) {
		largeSecret := strings.Repeat("secret", 10000)

		assert.NotPanics(t, func() {
			logger.Info("Large secret test",
				zap.String("password", largeSecret),
				zap.String("normal_field", "normal_value"),
			)
		})
	})

	t.Run("mixed_case_sensitive_fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Info("Mixed case test",
				zap.String("PASSWORD", "secret123"),
				zap.String("Secret", "secret456"),
				zap.String("TOKEN", "token789"),
				zap.String("normal", "value"),
			)
		})
	})
}

// TestLoggerCreationEdgeCases tests edge cases in logger creation
func TestLoggerCreationEdgeCases(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		logger, err := NewLogger(nil)
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("config_with_nil_log_level", func(t *testing.T) {
		// This might be a struct with uninitialized string field
		config := &SimpleConfig{
			Development: false,
			// LogLevel not set, should be empty string
		}

		logger, err := NewLogger(config)
		assert.Error(t, err)
		assert.Nil(t, logger)
	})

	t.Run("concurrent_logger_creation", func(t *testing.T) {
		config := &SimpleConfig{
			LogLevel:    "info",
			Development: false,
		}

		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		// Create multiple loggers concurrently
		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := NewLogger(config)
				results <- err
			}()
		}

		// All should succeed
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent logger creation should not fail")
		}
	})
}

// TestHighVolumeLoggingStress tests logger behavior under high volume
func TestHighVolumeLoggingStress(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	t.Run("rapid_sequential_logging", func(t *testing.T) {
		// Log many messages rapidly
		const numMessages = 1000

		assert.NotPanics(t, func() {
			for i := 0; i < numMessages; i++ {
				logger.Info("Rapid message",
					zap.Int("iteration", i),
					zap.String("data", "test-data"),
				)
			}
		})
	})

	t.Run("concurrent_high_volume_logging", func(t *testing.T) {
		const numGoroutines = 10
		const messagesPerGoroutine = 100
		done := make(chan bool, numGoroutines)

		// Launch multiple goroutines logging concurrently
		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for i := 0; i < messagesPerGoroutine; i++ {
					logger.Info("Concurrent message",
						zap.Int("goroutine", goroutineID),
						zap.Int("message", i),
						zap.String("password", "secret123"), // Should be sanitized
					)
				}
			}(g)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

// TestMemoryPressure tests logger behavior under memory pressure
func TestMemoryPressure(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	t.Run("large_log_messages", func(t *testing.T) {
		// Create very large log messages
		largeMessage := strings.Repeat("This is a very long log message. ", 1000)

		assert.NotPanics(t, func() {
			for i := 0; i < 10; i++ {
				logger.Info(largeMessage,
					zap.String("large_field", largeMessage),
					zap.Int("iteration", i),
				)
			}
		})
	})

	t.Run("many_fields_per_message", func(t *testing.T) {
		fields := make([]zap.Field, 100)
		for i := 0; i < 100; i++ {
			fields[i] = zap.String("field_"+string(rune(i)), "value_"+string(rune(i)))
		}

		assert.NotPanics(t, func() {
			logger.Info("Message with many fields", fields...)
		})
	})
}

// TestErrorRecovery tests logger's ability to recover from errors
func TestErrorRecovery(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	t.Run("continue_after_invalid_fields", func(t *testing.T) {
		// Log with potentially problematic fields
		assert.NotPanics(t, func() {
			logger.Info("Test message",
				zap.Any("nil_value", nil),
				zap.Any("func_value", func() {}), // Functions can't be JSON marshaled
				zap.String("normal_field", "normal_value"),
			)

			// Should continue working normally after error
			logger.Info("Follow-up message", zap.String("status", "ok"))
		})
	})
}
