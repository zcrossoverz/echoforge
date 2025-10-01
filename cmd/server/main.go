// cmd/server/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"github.com/zcrossoverz/echoforge/pkg/common"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize Viper configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.SetDefault("port", "8080")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("site_id", "default")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use defaults
	}

	// Setup logger with config level
	logger := common.WithConfig(viper.GetString("log_level"))
	defer logger.Sync() // Flush buffered logs

	// Generate unique server ID
	serverID := uuid.New().String()

	logger.Info("app starting",
		zap.String("server_id", serverID),
		zap.String("site_id", viper.GetString("site_id")),
		zap.String("port", viper.GetString("port")),
		zap.String("mode", viper.GetString("mode")),
	)

	// Demonstrate bcrypt usage
	testPassword := "test_password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Password hashing failed", zap.Error(err))
	} else {
		logger.Info("Password hashing successful", zap.Int("hash_length", len(hashedPassword)))
	}

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Basic health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"server_id": serverID,
			"site_id":   viper.GetString("site_id"),
		})
	})

	// Database connection placeholder
	dsn := viper.GetString("database.dsn")
	if dsn == "" {
		dsn = "host=localhost user=echoforge password=echoforge dbname=echoforge port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Info("Database connection failed (expected without DB setup)", zap.Error(err))
	} else {
		logger.Info("Database connected successfully")
		// Use db to prevent unused variable error
		_ = db
	}

	// Wire DI placeholder - in real app this would be used for dependency injection
	_ = wire.Build

	// Start server with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server in goroutine
	go func() {
		port := viper.GetString("port")
		logger.Info("HTTP server starting", zap.String("port", port))
		if err := router.Run(":" + port); err != nil {
			logger.Error("Server failed to start", zap.Error(err))
			cancel()
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	logger.Info("app shutting down gracefully")
	<-shutdownCtx.Done()
}
