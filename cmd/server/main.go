// cmd/server/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zcrossoverz/echoforge/pkg/common"
	"go.uber.org/zap"
)

func main() {
	// Load config
	cfg, err := LoadConfig("configs/config.yaml")
	if err != nil {
		panic(err) // Prod: Use logger.Fatal
	}

	// Setup logger với config level
	logger := common.WithConfig(cfg.App.LogLevel)
	defer logger.Sync() // Flush buffered logs

	logger.Info("app starting",
		zap.String("site_id", cfg.App.SiteID),
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	// TODO: Wire usecase/repo, start Gin server
	// Stub: Context timeout để simulate run
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("app shutting down gracefully")
}
