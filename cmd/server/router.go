package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/adapters/http"
	"go.uber.org/zap"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Logger        *zap.Logger
	AuthHandler   *http.AuthHandler
	HealthHandler *http.HealthHandler
	Environment   string
}

// SetupRouter configures and returns a Gin router with all routes and middleware
func SetupRouter(config *RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router with default middleware (Logger and Recovery)
	router := gin.New()

	// Global middleware stack
	router.Use(gin.Logger())   // HTTP request logging
	router.Use(gin.Recovery()) // Panic recovery

	// CORS middleware for cross-origin requests
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// Root health check (for load balancers)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":   "echoforge",
			"version":   "1.0.0",
			"status":    "ok",
			"endpoints": []string{"/api/v1/health", "/api/v1/auth/*"},
		})
	})

	// API versioning - v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoints (no authentication required)
		setupHealthRoutes(v1, config.HealthHandler)

		// Authentication endpoints
		setupAuthRoutes(v1, config.AuthHandler)
	}

	config.Logger.Info("Router configured successfully",
		zap.String("mode", gin.Mode()),
		zap.Int("route_count", len(router.Routes())),
	)

	return router
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(group *gin.RouterGroup, handler *http.HealthHandler) {
	health := group.Group("/health")
	{
		// Basic health check
		health.GET("", handler.HealthCheck)

		// Kubernetes-style readiness probe
		health.GET("/ready", handler.ReadinessCheck)

		// Kubernetes-style liveness probe
		health.GET("/live", handler.LivenessCheck)

		// Quick health check (for monitoring)
		health.GET("/quick", handler.QuickHealthCheck)
	}
}

// setupAuthRoutes configures authentication endpoints
func setupAuthRoutes(
	group *gin.RouterGroup,
	handler *http.AuthHandler,
) {
	auth := group.Group("/auth")

	// Public authentication endpoints (no auth required)
	{
		// User registration
		auth.POST("/register", handler.Register)

		// User login
		auth.POST("/login", handler.Login)
	}

	// Protected authentication endpoints (auth required)
	// Note: Authentication middleware will be applied via Wire DI
	{
		// User logout
		auth.POST("/logout", handler.Logout)

		// Get user profile
		auth.GET("/profile", handler.GetProfile)

		// Update user profile (future endpoint)
		auth.PUT("/profile", func(c *gin.Context) {
			c.JSON(501, gin.H{
				"success": false,
				"message": "Profile update not implemented yet",
				"error": gin.H{
					"code":    "NOT_IMPLEMENTED",
					"message": "This endpoint will be implemented in a future version",
				},
			})
		})

		// Delete user account (future endpoint)
		auth.DELETE("/profile", func(c *gin.Context) {
			c.JSON(501, gin.H{
				"success": false,
				"message": "Account deletion not implemented yet",
				"error": gin.H{
					"code":    "NOT_IMPLEMENTED",
					"message": "This endpoint will be implemented in a future version",
				},
			})
		})
	}
}

// RouteInfo represents information about a route
type RouteInfo struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Handler   string `json:"handler"`
	Protected bool   `json:"protected"`
}

// GetRouteInfo returns information about all configured routes
func GetRouteInfo(router *gin.Engine) []RouteInfo {
	routes := router.Routes()
	routeInfos := make([]RouteInfo, 0, len(routes))

	for _, route := range routes {
		info := RouteInfo{
			Method:    route.Method,
			Path:      route.Path,
			Handler:   route.Handler,
			Protected: isProtectedRoute(route.Path),
		}
		routeInfos = append(routeInfos, info)
	}

	return routeInfos
}

// isProtectedRoute checks if a route requires authentication
func isProtectedRoute(path string) bool {
	protectedPaths := []string{
		"/api/v1/auth/logout",
		"/api/v1/auth/profile",
	}

	for _, protectedPath := range protectedPaths {
		if path == protectedPath {
			return true
		}
	}

	return false
}

// ValidateRouterConfig validates router configuration
func ValidateRouterConfig(config *RouterConfig) error {
	if config.Logger == nil {
		return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate}
	}

	if config.AuthHandler == nil {
		return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate}
	}

	if config.HealthHandler == nil {
		return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate}
	}

	return nil
}
