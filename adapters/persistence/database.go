package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zcrossoverz/echoforge/internal/config"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Database represents the database connection and configuration
type Database struct {
	DB     *gorm.DB
	Config *DatabaseConfig
}

// NewDatabase creates a new database connection using DSN from config
func NewDatabase(cfg *config.Config) (*Database, error) {
	// Set default connection pool settings
	dbConfig := &DatabaseConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	db, err := connectPostgreSQLWithDSN(cfg.DBDSN, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{
		DB:     db,
		Config: dbConfig,
	}, nil
}

// connectPostgreSQLWithDSN establishes a connection to PostgreSQL database using DSN
func connectPostgreSQLWithDSN(dsn string, config *DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	return db, nil
}

// Ping tests the database connection
func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

// AutoMigrate runs auto-migration for all models
func (d *Database) AutoMigrate() error {
	// Run auto-migration for all models
	err := d.DB.AutoMigrate(
		&GormUser{},
		&AuthBlacklistModel{},
	)

	if err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func (d *Database) GetStats() (*DatabaseStats, error) {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	return &DatabaseStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// DatabaseStats represents database connection statistics
type DatabaseStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// AuthBlacklistModel represents the GORM model for auth_blacklist table
type AuthBlacklistModel struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TokenHash     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"token_hash"`
	ExpiresAt     time.Time `gorm:"not null;index" json:"expires_at"`
	BlacklistedAt time.Time `gorm:"autoCreateTime" json:"blacklisted_at"`
	UserID        *string   `gorm:"type:uuid;index" json:"user_id"` // Nullable for cleanup purposes
}

// TableName specifies the table name for GORM
func (AuthBlacklistModel) TableName() string {
	return "auth_blacklist"
}

// HealthCheck performs a comprehensive health check of the database
func (d *Database) HealthCheck(ctx context.Context) error {
	// Test basic connectivity
	if err := d.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test with a simple query
	var result int
	if err := d.DB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	return nil
}

// IsHealthy checks if the database is healthy (non-blocking)
func (d *Database) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.HealthCheck(ctx) == nil
}
