//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/zcrossoverz/echoforge/adapters/http"
	"github.com/zcrossoverz/echoforge/adapters/http/middleware"
	"github.com/zcrossoverz/echoforge/adapters/persistence"
	"github.com/zcrossoverz/echoforge/internal/config"
	"github.com/zcrossoverz/echoforge/internal/logging"
	"github.com/zcrossoverz/echoforge/internal/usecase"
	"github.com/zcrossoverz/echoforge/pkg/auth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application represents the main application container
type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	DB            *gorm.DB
	Router        *gin.Engine
	AuthHandler   *http.AuthHandler
	HealthHandler *http.HealthHandler
}

// Wire provider sets

// ConfigProvider provides configuration
var ConfigProvider = wire.NewSet(
	config.NewConfig,
)

// LoggingProvider provides logging
var LoggingProvider = wire.NewSet(
	NewLoggingConfig,
	logging.NewLogger,
)

// DatabaseProvider provides database connection
var DatabaseProvider = wire.NewSet(
	NewDatabaseConnection,
)

// AuthProvider provides authentication services
var AuthProvider = wire.NewSet(
	auth.NewJWTService,
	auth.NewPasswordService,
)

// RepositoryProvider provides repositories
var RepositoryProvider = wire.NewSet(
	persistence.NewUserRepository,
	wire.Bind(new(domain.UserRepository), new(*persistence.UserRepository)),
)

// UseCaseProvider provides use cases
var UseCaseProvider = wire.NewSet(
	usecase.NewUserUseCase,
	usecase.NewUserRegistrationUseCase,
	usecase.NewUserAuthenticationUseCase,
	usecase.NewUserLogoutUseCase,
	usecase.NewGetUserProfileUseCase,
)

// MiddlewareProvider provides middleware
var MiddlewareProvider = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewValidationMiddleware,
)

// HandlerProvider provides HTTP handlers
var HandlerProvider = wire.NewSet(
	http.NewAuthHandler,
	http.NewHealthHandler,
)

// RouterProvider provides router
var RouterProvider = wire.NewSet(
	NewSimpleRouter,
)

// ApplicationProvider provides the complete application
var ApplicationProvider = wire.NewSet(
	ConfigProvider,
	LoggingProvider,
	DatabaseProvider,
	AuthProvider,
	RepositoryProvider,
	UseCaseProvider,
	MiddlewareProvider,
	HandlerProvider,
	RouterProvider,
	NewApplication,
)

// InitializeApplication creates a fully configured application
func InitializeApplication() (*Application, error) {
	wire.Build(ApplicationProvider)
	return &Application{}, nil
}

// Provider functions

// NewLoggingConfig creates logging configuration from app config
func NewLoggingConfig(cfg *config.Config) *logging.SimpleConfig {
	return &logging.SimpleConfig{
		LogLevel:    cfg.LogLevel,
		Development: cfg.EnableHotReload, // Use hot reload as development indicator
	}
}

// NewDatabaseConnection creates database connection from config
func NewDatabaseConnection(cfg *config.Config) (*gorm.DB, error) {
	database, err := persistence.NewDatabase(cfg)
	if err != nil {
		return nil, err
	}
	return database.DB, nil
}

// NewSimpleRouter creates a simple router with basic setup
func NewSimpleRouter(
	logger *zap.Logger,
	authHandler *http.AuthHandler,
	healthHandler *http.HealthHandler,
) *gin.Engine {
	config := &RouterConfig{
		Logger:        logger,
		AuthHandler:   authHandler,
		HealthHandler: healthHandler,
		Environment:   "production",
	}

	return SetupRouter(config)
}

// NewApplication creates the main application instance
func NewApplication(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	router *gin.Engine,
	authHandler *http.AuthHandler,
	healthHandler *http.HealthHandler,
) *Application {
	return &Application{
		Config:        cfg,
		Logger:        logger,
		DB:            db,
		Router:        router,
		AuthHandler:   authHandler,
		HealthHandler: healthHandler,
	}
}
