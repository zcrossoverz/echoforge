package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zcrossoverz/echoforge/adapters/persistence"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	database *persistence.Database
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(database *persistence.Database) *HealthHandler {
	return &HealthHandler{
		database: database,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
	Uptime    string                 `json:"uptime"`
}

// ServiceInfo represents individual service health information
type ServiceInfo struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

// HealthCheck handles GET /api/v1/health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	startTime := time.Now()

	// Initialize health response
	healthResponse := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		Version:   "1.0.0", // TODO: Get from build info or config
		Services:  make(map[string]ServiceInfo),
		Uptime:    getUptime().String(),
	}

	// Check database health
	dbStatus := h.checkDatabaseHealth()
	healthResponse.Services["database"] = dbStatus

	// Determine overall health status
	if dbStatus.Status != "healthy" {
		healthResponse.Status = "unhealthy"
	}

	// Calculate total response time
	responseTime := time.Since(startTime)

	// Add response time to health info
	healthResponse.Services["api"] = ServiceInfo{
		Status:       "healthy",
		ResponseTime: responseTime.String(),
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if healthResponse.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, healthResponse)
}

// checkDatabaseHealth checks the database connectivity and performance
func (h *HealthHandler) checkDatabaseHealth() ServiceInfo {
	if h.database == nil {
		return ServiceInfo{
			Status: "unhealthy",
			Error:  "database not configured",
		}
	}

	startTime := time.Now()

	// Create a context with timeout for the health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform database health check
	if err := h.database.HealthCheck(ctx); err != nil {
		return ServiceInfo{
			Status:       "unhealthy",
			ResponseTime: time.Since(startTime).String(),
			Error:        err.Error(),
		}
	}

	return ServiceInfo{
		Status:       "healthy",
		ResponseTime: time.Since(startTime).String(),
	}
}

// getUptime calculates the application uptime
// This is a simple implementation - in production, you might want to track
// the actual server start time
func getUptime() time.Duration {
	// For now, return a placeholder uptime
	// In a real application, you would track the server start time
	return time.Since(time.Now().Add(-1 * time.Hour)) // Placeholder: 1 hour uptime
}

// QuickHealthCheck provides a lightweight health check without detailed service checks
func (h *HealthHandler) QuickHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ReadinessCheck checks if the application is ready to serve traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check critical dependencies
	if h.database == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"reason":    "database not configured",
			"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
		return
	}

	// Quick database connectivity test
	if !h.database.IsHealthy() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"reason":    "database not healthy",
			"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// LivenessCheck checks if the application is alive (basic functionality)
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}
