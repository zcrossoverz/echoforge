// pkg/common/logger.go
package common

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(output zapcore.WriteSyncer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()) // Structured JSON, light

	core := zapcore.NewCore(encoder, output, zapcore.InfoLevel) // Default level
	return zap.New(core, zap.AddCaller())                       // Add caller for stack
}

func DefaultLogger() *zap.Logger {
	return NewLogger(zapcore.AddSync(os.Stdout))
}

// WithConfig: Dynamic level từ Viper (atomic swap)
func WithConfig(level string) *zap.Logger {
	lvl, _ := zapcore.ParseLevel(level) // e.g., "debug", "info"
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(lvl))
	return zap.New(core, zap.AddCaller())
}
