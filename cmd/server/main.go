package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	// Initialize application using Wire dependency injection
	app, err := InitializeApplication()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	// Generate unique server ID for tracking
	serverID := uuid.New().String()

	app.Logger.Info("Echoforge server starting",
		zap.String("server_id", serverID),
		zap.String("log_level", app.Config.LogLevel),
		zap.String("environment", getEnvironment()),
		zap.String("version", "1.0.0"),
	)

	// Validate configuration
	if err := ValidateServerConfig(app.Config); err != nil {
		app.Logger.Fatal("Configuration validation failed", zap.Error(err))
	}

	// Setup HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", getPort()),
		Handler:        app.Router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		app.Logger.Info("HTTP server starting",
			zap.String("address", server.Addr),
			zap.String("server_id", serverID),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("Server failed to start", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	gracefulShutdown(ctx, cancel, server, app.Logger)
}

// gracefulShutdown handles graceful server shutdown
func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server, logger *zap.Logger) {
	// Create channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or context cancellation
	select {
	case sig := <-sigChan:
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
	case <-ctx.Done():
		logger.Info("Application context cancelled")
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("Initiating graceful shutdown...")

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))

		// Force close if graceful shutdown fails
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("Server force close failed", zap.Error(closeErr))
		}
	} else {
		logger.Info("Server shutdown completed successfully")
	}
}

// getPort returns the port from environment or default
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// getEnvironment returns the environment from ENV or default
func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return env
}
