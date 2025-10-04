package logging

import (
	"context"
	"runtime"
	"testing"
	"time"

	"go.uber.org/zap"
)

// BenchmarkNewLogger benchmarks logger creation performance
func BenchmarkNewLogger(b *testing.B) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger, err := NewLogger(config)
		if err != nil {
			b.Fatalf("NewLogger failed: %v", err)
		}
		_ = logger
	}
}

// BenchmarkLoggingThroughput benchmarks logging throughput
func BenchmarkLoggingThroughput(b *testing.B) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("NewLogger failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message",
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
			zap.Int("number", i),
		)
	}
}

// BenchmarkContextualLogging benchmarks context-aware logging
func BenchmarkContextualLogging(b *testing.B) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("NewLogger failed: %v", err)
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123456789")
	ctx = WithUserID(ctx, "user_123")

	contextLogger := NewContextualLogger(logger, ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		contextLogger.Info("Benchmark message",
			zap.String("key1", "value1"),
			zap.Int("iteration", i),
		)
	}
}

// BenchmarkSanitizedLogging benchmarks logging with sanitization
func BenchmarkSanitizedLogging(b *testing.B) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("NewLogger failed: %v", err)
	}

	sanitizedLogger := CreateSanitizingLogger(logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sanitizedLogger.Info("Benchmark message",
			zap.String("username", "testuser"),
			zap.String("password", "secret123"),  // Will be sanitized
			zap.String("token", "jwt-token-123"), // Will be sanitized
			zap.Int("iteration", i),
		)
	}
}

// BenchmarkLoggerTypes benchmarks different logger configurations
func BenchmarkLoggerTypes(b *testing.B) {
	tests := []struct {
		name   string
		config *SimpleConfig
	}{
		{
			name: "Development",
			config: &SimpleConfig{
				LogLevel:    "debug",
				Development: true,
			},
		},
		{
			name: "Production",
			config: &SimpleConfig{
				LogLevel:    "info",
				Development: false,
			},
		},
		{
			name: "ErrorOnly",
			config: &SimpleConfig{
				LogLevel:    "error",
				Development: false,
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			logger, err := NewLogger(tt.config)
			if err != nil {
				b.Fatalf("NewLogger failed: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger.Info("Benchmark message",
					zap.String("type", tt.name),
					zap.Int("iteration", i),
				)
			}
		})
	}
}

// BenchmarkSensitiveFieldFiltering benchmarks field filtering performance
func BenchmarkSensitiveFieldFiltering(b *testing.B) {
	tests := []struct {
		name   string
		filter SensitiveFieldFilter
	}{
		{
			name:   "Default",
			filter: NewDefaultSensitiveFieldFilter(),
		},
		{
			name:   "Enhanced",
			filter: NewEnhancedSensitiveFieldFilter(),
		},
	}

	testFields := map[string]interface{}{
		"username":     "testuser",
		"password":     "secret123",
		"token":        "jwt-token-123",
		"dsn":          "postgres://user:pass@localhost/db",
		"normal_field": "normal_value",
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for key, value := range testFields {
					_ = tt.filter.Sanitize(key, value)
				}
			}
		})
	}
}

// BenchmarkConcurrentLogging benchmarks concurrent logging performance
func BenchmarkConcurrentLogging(b *testing.B) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("NewLogger failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("Concurrent benchmark message",
				zap.Int("goroutine_id", i%1000),
				zap.String("message", "test message"),
			)
			i++
		}
	})
}

// Performance tests with specific requirements

// TestLoggingThroughput tests that logging can handle 1000+ logs/sec
func TestLoggingThroughput(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	const targetLogs = 1000
	const testDuration = 1 * time.Second

	start := time.Now()
	count := 0

	for time.Since(start) < testDuration {
		logger.Info("Throughput test message",
			zap.Int("count", count),
			zap.String("key", "value"),
		)
		count++
	}

	duration := time.Since(start)
	logsPerSecond := float64(count) / duration.Seconds()

	t.Logf("Logged %d messages in %v (%.2f logs/sec)", count, duration, logsPerSecond)

	if logsPerSecond < 1000 {
		t.Errorf("Logging throughput too low: %.2f logs/sec, expected >= 1000", logsPerSecond)
	}
}

// TestLoggingMemoryFootprint tests memory usage stays under 50MB
func TestLoggingMemoryFootprint(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Get baseline memory
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)

	// Log many messages
	const numLogs = 10000
	for i := 0; i < numLogs; i++ {
		logger.Info("Memory test message",
			zap.Int("iteration", i),
			zap.String("data", "some test data here"),
			zap.Time("timestamp", time.Now()),
		)
	}

	// Force garbage collection and measure
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)

	memoryUsed := final.Alloc - baseline.Alloc
	memoryUsedMB := float64(memoryUsed) / (1024 * 1024)

	t.Logf("Memory used for %d log messages: %.2f MB", numLogs, memoryUsedMB)

	if memoryUsedMB > 50 {
		t.Errorf("Memory footprint too high: %.2f MB, expected <= 50 MB", memoryUsedMB)
	}
}

// TestContextOverhead tests that context propagation doesn't add significant overhead
func TestContextOverhead(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Benchmark regular logging
	regularResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("Regular message", zap.Int("i", i))
		}
	})

	// Benchmark contextual logging
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123456789")
	ctx = WithUserID(ctx, "user_123")
	contextLogger := NewContextualLogger(logger, ctx)

	contextResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			contextLogger.Info("Context message", zap.Int("i", i))
		}
	})

	regularNsPerOp := regularResult.NsPerOp()
	contextNsPerOp := contextResult.NsPerOp()
	overhead := float64(contextNsPerOp-regularNsPerOp) / float64(regularNsPerOp) * 100

	t.Logf("Regular logging: %d ns/op", regularNsPerOp)
	t.Logf("Context logging: %d ns/op", contextNsPerOp)
	t.Logf("Context overhead: %.2f%%", overhead)

	// Context overhead should be reasonable (less than 50%)
	if overhead > 50 {
		t.Errorf("Context overhead too high: %.2f%%, expected <= 50%%", overhead)
	}
}

// Stress tests

// TestHighVolumeLogging tests logging under high volume
func TestHighVolumeLogging(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	const numGoroutines = 10
	const logsPerGoroutine = 1000

	start := time.Now()
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < logsPerGoroutine; i++ {
				logger.Info("High volume test",
					zap.Int("goroutine", goroutineID),
					zap.Int("iteration", i),
					zap.String("data", "test data"),
				)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	duration := time.Since(start)
	totalLogs := numGoroutines * logsPerGoroutine
	logsPerSecond := float64(totalLogs) / duration.Seconds()

	t.Logf("High volume test: %d logs in %v (%.2f logs/sec)", totalLogs, duration, logsPerSecond)

	// Should maintain reasonable performance under load
	if logsPerSecond < 500 {
		t.Errorf("High volume performance too low: %.2f logs/sec", logsPerSecond)
	}
}

// Memory allocation tests

// TestLoggingAllocations tests memory allocations per log operation
func TestLoggingAllocations(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Run benchmark to measure allocations
	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("Allocation test message",
				zap.String("key1", "value1"),
				zap.Int("key2", i),
			)
		}
	})

	allocsPerOp := result.MemAllocs / uint64(result.N)
	bytesPerOp := result.MemBytes / uint64(result.N)

	t.Logf("Logging allocations: %d allocs/op, %d bytes/op", allocsPerOp, bytesPerOp)

	// Reasonable limits for logging operations
	if allocsPerOp > 10 {
		t.Errorf("Too many allocations per log: %d", allocsPerOp)
	}

	if bytesPerOp > 1024 { // 1KB per log operation
		t.Errorf("Too much memory per log: %d bytes", bytesPerOp)
	}
}

// TestSanitizationOverhead tests the performance impact of sanitization
func TestSanitizationOverhead(t *testing.T) {
	config := &SimpleConfig{
		LogLevel:    "info",
		Development: false,
	}

	regularLogger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	sanitizedLogger := CreateSanitizingLogger(regularLogger)

	// Benchmark regular logging
	regularResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			regularLogger.Info("Test message",
				zap.String("password", "secret123"),
				zap.String("normal", "value"),
			)
		}
	})

	// Benchmark sanitized logging
	sanitizedResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sanitizedLogger.Info("Test message",
				zap.String("password", "secret123"),
				zap.String("normal", "value"),
			)
		}
	})

	regularNsPerOp := regularResult.NsPerOp()
	sanitizedNsPerOp := sanitizedResult.NsPerOp()
	overhead := float64(sanitizedNsPerOp-regularNsPerOp) / float64(regularNsPerOp) * 100

	t.Logf("Regular logging: %d ns/op", regularNsPerOp)
	t.Logf("Sanitized logging: %d ns/op", sanitizedNsPerOp)
	t.Logf("Sanitization overhead: %.2f%%", overhead)

	// Sanitization overhead should be reasonable (less than 100%)
	if overhead > 100 {
		t.Errorf("Sanitization overhead too high: %.2f%%, expected <= 100%%", overhead)
	}
}
