package config

import (
	"os"
	"testing"
	"time"
)

// BenchmarkNewConfig benchmarks the config loading performance
func BenchmarkNewConfig(b *testing.B) {
	// Clean up environment to ensure consistent benchmarks
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	// Set test environment variables
	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := NewConfig()
		if err != nil {
			b.Fatalf("NewConfig failed: %v", err)
		}
	}
}

// BenchmarkConfigLoading tests config loading under different scenarios
func BenchmarkConfigLoading(b *testing.B) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "EnvOnly",
			setup: func() {
				os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
				os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
				os.Setenv("LOG_LEVEL", "info")
				os.Setenv("ENABLE_HOT_RELOAD", "false")
			},
		},
		{
			name: "WithDefaults",
			setup: func() {
				os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
				os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
				// Let LOG_LEVEL use default
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("ENABLE_HOT_RELOAD")
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tt.setup()
			defer func() {
				os.Unsetenv("DB_DSN")
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("ENABLE_HOT_RELOAD")
			}()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := NewConfig()
				if err != nil {
					b.Fatalf("NewConfig failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkHotReloadSetup benchmarks the hot-reload setup performance
func BenchmarkHotReloadSetup(b *testing.B) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "true")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		config, watcher, err := HotReloadConfig()
		if err != nil {
			b.Fatalf("HotReloadConfig failed: %v", err)
		}

		// Clean up
		if watcher != nil {
			watcher.Stop()
		}
		_ = config
	}
}

// BenchmarkConfigWithWatcher benchmarks the ConfigWithWatcher creation
func BenchmarkConfigWithWatcher(b *testing.B) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false") // Disable for faster benchmarks

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		configWrapper, err := NewConfigWithHotReload()
		if err != nil {
			b.Fatalf("NewConfigWithHotReload failed: %v", err)
		}

		// Clean up
		configWrapper.Close()
	}
}

// BenchmarkLegacyConfigAccess benchmarks legacy config access patterns
func BenchmarkLegacyConfigAccess(b *testing.B) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	// Initialize legacy config
	cfg, err := NewConfig()
	if err != nil {
		b.Fatalf("NewConfig failed: %v", err)
	}
	InitLegacyConfig(cfg)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		legacy := GetLegacyConfig()
		_ = legacy.GetDBDSN()
		_ = legacy.GetJWTSecret()
		_ = legacy.GetLogLevel()
		_ = legacy.IsHotReloadEnabled()
	}
}

// BenchmarkConfigValidation benchmarks the validation performance
func BenchmarkConfigValidation(b *testing.B) {
	configs := []*Config{
		{
			DBDSN:           "postgres://user:pass@localhost/test",
			JWTSecret:       "test-secret-key-that-is-32-chars-long!",
			LogLevel:        "info",
			EnableHotReload: false,
		},
		{
			DBDSN:           "postgres://user:pass@localhost/test",
			JWTSecret:       "test-secret-key-that-is-32-chars-long!",
			LogLevel:        "debug",
			EnableHotReload: true,
		},
		{
			DBDSN:           "postgres://user:pass@localhost/test",
			JWTSecret:       "test-secret-key-that-is-32-chars-long!",
			LogLevel:        "error",
			EnableHotReload: false,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, cfg := range configs {
			err := validate.Struct(cfg)
			if err != nil {
				b.Fatalf("Validation failed: %v", err)
			}
		}
	}
}

// Performance tests with time constraints

// TestConfigLoadingPerformance tests that config loading is under 5 seconds
func TestConfigLoadingPerformance(t *testing.T) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	start := time.Now()
	_, err := NewConfig()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if duration > 5*time.Second {
		t.Errorf("Config loading took %v, expected < 5s", duration)
	}

	t.Logf("Config loading took %v", duration)
}

// TestHotReloadPerformance tests that hot-reload setup is under 1 second
func TestHotReloadPerformance(t *testing.T) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "true")

	start := time.Now()
	config, watcher, err := HotReloadConfig()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("HotReloadConfig failed: %v", err)
	}

	defer func() {
		if watcher != nil {
			watcher.Stop()
		}
	}()

	if duration > 1*time.Second {
		t.Errorf("Hot-reload setup took %v, expected < 1s", duration)
	}

	t.Logf("Hot-reload setup took %v", duration)
	_ = config
}

// Memory usage tests

// TestConfigMemoryUsage tests memory usage of config operations
func TestConfigMemoryUsage(t *testing.T) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	// Run a benchmark to measure memory allocations
	result := testing.Benchmark(BenchmarkNewConfig)

	// Check memory allocations per operation
	allocsPerOp := result.MemAllocs / uint64(result.N)
	bytesPerOp := result.MemBytes / uint64(result.N)

	t.Logf("Config loading: %d allocs/op, %d bytes/op", allocsPerOp, bytesPerOp)

	// Reasonable limits for config loading (adjusted for Viper overhead)
	if allocsPerOp > 500 {
		t.Errorf("Too many allocations per config load: %d", allocsPerOp)
	}

	if bytesPerOp > 50*1024 { // 50KB per config load (Viper is heavy)
		t.Errorf("Too much memory per config load: %d bytes", bytesPerOp)
	}
}

// Stress test for concurrent config access
func TestConcurrentConfigAccess(t *testing.T) {
	// Set test environment
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENABLE_HOT_RELOAD")
	}()

	os.Setenv("DB_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-32-chars-long!")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENABLE_HOT_RELOAD", "false")

	// Initialize legacy config
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}
	InitLegacyConfig(cfg)

	// Run concurrent access test
	const numGoroutines = 100
	const numOperations = 1000

	done := make(chan bool, numGoroutines)

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		go func() {
			legacy := GetLegacyConfig()
			for j := 0; j < numOperations; j++ {
				_ = legacy.GetDBDSN()
				_ = legacy.GetJWTSecret()
				_ = legacy.GetLogLevel()
				_ = legacy.IsHotReloadEnabled()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	duration := time.Since(start)
	t.Logf("Concurrent access test completed in %v", duration)

	// Should complete in reasonable time
	if duration > 5*time.Second {
		t.Errorf("Concurrent access took too long: %v", duration)
	}
}
