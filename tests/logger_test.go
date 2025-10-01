// tests/logger_test.go
package tests

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zcrossoverz/echoforge/pkg/common"
	"go.uber.org/zap"
)

func TestLogger_Info(t *testing.T) {
	// Capture output với JSON encoder (Zap production mode)
	buf := new(bytes.Buffer)
	logger := common.NewLogger(buf)

	// Log structured
	logger.Info("user registered", zap.String("site_id", "blog1"))

	// Verify JSON output
	output := buf.String()
	assert.Contains(t, output, `"msg":"user registered"`)
	assert.Contains(t, output, `"site_id":"blog1"`)
	assert.Contains(t, output, `"level":"info"`) // Level check
}

func TestDefaultLogger(t *testing.T) {
	logger := common.DefaultLogger()
	assert.NotNil(t, logger)
	assert.NotPanics(t, func() { logger.Info("test") }) // Basic functional
}
