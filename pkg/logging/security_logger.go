package logging

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SecurityLogger provides specialized logging for security events
type SecurityLogger struct {
	logger *zap.Logger
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	ClientIP  string                 `json:"client_ip"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Success   bool                   `json:"success"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Severity  string                 `json:"severity"`
	RequestID string                 `json:"request_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// SecurityEventType constants for different types of security events
const (
	EventTypeLogin              = "LOGIN"
	EventTypeLoginFailed        = "LOGIN_FAILED"
	EventTypeLogout             = "LOGOUT"
	EventTypeRegistration       = "REGISTRATION"
	EventTypePasswordChange     = "PASSWORD_CHANGE"
	EventTypeAccountLocked      = "ACCOUNT_LOCKED"
	EventTypeTokenExpired       = "TOKEN_EXPIRED"
	EventTypeTokenBlacklisted   = "TOKEN_BLACKLISTED"
	EventTypeUnauthorizedAccess = "UNAUTHORIZED_ACCESS"
	EventTypeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	EventTypeSuspiciousActivity = "SUSPICIOUS_ACTIVITY"
	EventTypeDataAccess         = "DATA_ACCESS"
	EventTypeDataModification   = "DATA_MODIFICATION"
	EventTypeAPIKeyUsage        = "API_KEY_USAGE"
	EventTypeSecurityAlert      = "SECURITY_ALERT"
)

// Severity levels for security events
const (
	SeverityInfo     = "INFO"
	SeverityWarning  = "WARNING"
	SeverityError    = "ERROR"
	SeverityCritical = "CRITICAL"
)

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(baseLogger *zap.Logger) *SecurityLogger {
	// Create a specialized logger for security events
	securityLogger := baseLogger.With(
		zap.String("component", "security"),
		zap.String("log_type", "security_audit"),
	)

	return &SecurityLogger{
		logger: securityLogger,
	}
}

// LogSecurityEvent logs a security event with structured data
func (sl *SecurityLogger) LogSecurityEvent(event SecurityEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	fields := []zap.Field{
		zap.String("event_type", event.EventType),
		zap.Time("timestamp", event.Timestamp),
		zap.String("client_ip", event.ClientIP),
		zap.Bool("success", event.Success),
		zap.String("severity", event.Severity),
		zap.String("message", event.Message),
	}

	// Add optional fields
	if event.UserID != "" {
		fields = append(fields, zap.String("user_id", event.UserID))
	}
	if event.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", event.UserAgent))
	}
	if event.Resource != "" {
		fields = append(fields, zap.String("resource", event.Resource))
	}
	if event.Action != "" {
		fields = append(fields, zap.String("action", event.Action))
	}
	if event.ErrorCode != "" {
		fields = append(fields, zap.String("error_code", event.ErrorCode))
	}
	if event.RequestID != "" {
		fields = append(fields, zap.String("request_id", event.RequestID))
	}
	if event.SessionID != "" {
		fields = append(fields, zap.String("session_id", event.SessionID))
	}
	if event.Metadata != nil {
		fields = append(fields, zap.Any("metadata", event.Metadata))
	}

	// Log at appropriate level based on severity
	switch event.Severity {
	case SeverityInfo:
		sl.logger.Info("Security event", fields...)
	case SeverityWarning:
		sl.logger.Warn("Security event", fields...)
	case SeverityError:
		sl.logger.Error("Security event", fields...)
	case SeverityCritical:
		sl.logger.Error("CRITICAL Security event", fields...)
	default:
		sl.logger.Info("Security event", fields...)
	}
}

// LogAuthentication logs authentication events (login, logout, registration)
func (sl *SecurityLogger) LogAuthentication(ctx context.Context, eventType string, userID string, clientIP string, success bool, errorCode string, metadata map[string]interface{}) {
	severity := SeverityInfo
	if !success {
		severity = SeverityWarning
		if eventType == EventTypeLoginFailed {
			// Multiple failed login attempts might indicate brute force
			severity = SeverityError
		}
	}

	message := "Authentication event"
	if !success {
		message = "Authentication failed"
	}

	event := SecurityEvent{
		EventType: eventType,
		UserID:    userID,
		ClientIP:  clientIP,
		Success:   success,
		ErrorCode: errorCode,
		Message:   message,
		Severity:  severity,
		Metadata:  metadata,
	}

	// Add request context if available
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			event.RequestID = requestID
		}
		if sessionID, ok := ctx.Value("session_id").(string); ok {
			event.SessionID = sessionID
		}
	}

	sl.LogSecurityEvent(event)
}

// LogUnauthorizedAccess logs attempts to access protected resources without proper authorization
func (sl *SecurityLogger) LogUnauthorizedAccess(ctx context.Context, clientIP string, resource string, userAgent string, reason string) {
	event := SecurityEvent{
		EventType: EventTypeUnauthorizedAccess,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Resource:  resource,
		Success:   false,
		Message:   "Unauthorized access attempt",
		Severity:  SeverityWarning,
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}

	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			event.RequestID = requestID
		}
	}

	sl.LogSecurityEvent(event)
}

// LogRateLimitExceeded logs rate limiting events
func (sl *SecurityLogger) LogRateLimitExceeded(clientIP string, resource string, limit int, window time.Duration) {
	event := SecurityEvent{
		EventType: EventTypeRateLimitExceeded,
		ClientIP:  clientIP,
		Resource:  resource,
		Success:   false,
		Message:   "Rate limit exceeded",
		Severity:  SeverityWarning,
		Metadata: map[string]interface{}{
			"limit":  limit,
			"window": window.String(),
		},
	}

	sl.LogSecurityEvent(event)
}

// LogSuspiciousActivity logs potentially suspicious activities
func (sl *SecurityLogger) LogSuspiciousActivity(ctx context.Context, userID string, clientIP string, activity string, details map[string]interface{}) {
	event := SecurityEvent{
		EventType: EventTypeSuspiciousActivity,
		UserID:    userID,
		ClientIP:  clientIP,
		Success:   false,
		Message:   "Suspicious activity detected",
		Severity:  SeverityError,
		Metadata: map[string]interface{}{
			"activity": activity,
			"details":  details,
		},
	}

	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			event.RequestID = requestID
		}
	}

	sl.LogSecurityEvent(event)
}

// LogDataAccess logs access to sensitive data
func (sl *SecurityLogger) LogDataAccess(ctx context.Context, userID string, resource string, action string, success bool) {
	severity := SeverityInfo
	if !success {
		severity = SeverityWarning
	}

	event := SecurityEvent{
		EventType: EventTypeDataAccess,
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Success:   success,
		Message:   "Data access event",
		Severity:  severity,
	}

	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			event.RequestID = requestID
		}
		if clientIP, ok := ctx.Value("client_ip").(string); ok {
			event.ClientIP = clientIP
		}
	}

	sl.LogSecurityEvent(event)
}

// SecurityLoggingMiddleware provides Gin middleware for automatic security logging
func (sl *SecurityLogger) SecurityLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log security-relevant HTTP events
		if param.StatusCode >= 400 {
			severity := SeverityInfo
			if param.StatusCode >= 500 {
				severity = SeverityError
			} else if param.StatusCode >= 400 {
				severity = SeverityWarning
			}

			event := SecurityEvent{
				EventType: "HTTP_REQUEST",
				Timestamp: param.TimeStamp,
				ClientIP:  param.ClientIP,
				Resource:  param.Path,
				Action:    param.Method,
				Success:   param.StatusCode < 400,
				Message:   "HTTP request processed",
				Severity:  severity,
				Metadata: map[string]interface{}{
					"status_code":   param.StatusCode,
					"response_time": param.Latency,
					"response_size": param.BodySize,
					"user_agent":    param.Request.UserAgent(),
				},
			}

			sl.LogSecurityEvent(event)
		}

		// Return empty string as we're handling logging ourselves
		return ""
	})
}

// AuthenticationEventsMiddleware logs authentication-related events automatically
func (sl *SecurityLogger) AuthenticationEventsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log authentication attempts for auth endpoints
		if strings.Contains(c.Request.URL.Path, "/auth/") {
			startTime := time.Now()

			c.Next()

			duration := time.Since(startTime)
			success := c.Writer.Status() < 400

			var eventType string
			switch {
			case strings.Contains(c.Request.URL.Path, "/login"):
				eventType = EventTypeLogin
				if !success {
					eventType = EventTypeLoginFailed
				}
			case strings.Contains(c.Request.URL.Path, "/logout"):
				eventType = EventTypeLogout
			case strings.Contains(c.Request.URL.Path, "/register"):
				eventType = EventTypeRegistration
			default:
				return // Not an auth endpoint we care about
			}

			// Extract user ID from context if available
			userID := ""
			if userIDValue, exists := c.Get("user_id"); exists {
				if id, ok := userIDValue.(string); ok {
					userID = id
				}
			}

			metadata := map[string]interface{}{
				"response_time": duration,
				"status_code":   c.Writer.Status(),
				"user_agent":    c.Request.UserAgent(),
			}

			sl.LogAuthentication(
				c.Request.Context(),
				eventType,
				userID,
				c.ClientIP(),
				success,
				"", // Error code would be extracted from response
				metadata,
			)
		} else {
			c.Next()
		}
	}
}

// CreateSecurityAlertLogger creates a logger specifically for security alerts
func CreateSecurityAlertLogger() (*SecurityLogger, error) {
	// Create a high-priority logger for security alerts
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel) // Only log warnings and above
	loggerConfig.OutputPaths = []string{"stdout", "security_alerts.log"}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return NewSecurityLogger(logger), nil
}

// Helper function to add request context for security logging
func AddSecurityContext(c *gin.Context) {
	// Add request ID for tracing
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}
	c.Set("request_id", requestID)

	// Add client IP to context
	c.Set("client_ip", c.ClientIP())

	// Add session ID if available
	if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
		c.Set("session_id", sessionID)
	}
}

// generateRequestID generates a simple request ID for tracing
func generateRequestID() string {
	return time.Now().Format("20060102-150405") + "-" +
		time.Now().Format("000000")
}
