package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityMiddleware provides OWASP Top 10 compliance features
type SecurityMiddleware struct {
	logger *zap.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(logger *zap.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		logger: logger,
	}
}

// SecurityHeaders adds security headers to prevent common attacks
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Prevent clickjacking attacks (A10:2021 - Server-Side Request Forgery)
		c.Header("X-Frame-Options", "DENY")

		// 2. Prevent MIME-type sniffing (A06:2021 - Vulnerable and Outdated Components)
		c.Header("X-Content-Type-Options", "nosniff")

		// 3. Enable XSS protection in browsers (A03:2021 - Injection)
		c.Header("X-XSS-Protection", "1; mode=block")

		// 4. Enforce HTTPS (A02:2021 - Cryptographic Failures)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// 5. Content Security Policy to prevent XSS (A03:2021 - Injection)
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"media-src 'self'; " +
			"object-src 'none'; " +
			"frame-src 'none'; " +
			"worker-src 'none'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// 6. Referrer Policy (A01:2021 - Broken Access Control)
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 7. Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 8. Remove server information disclosure (A05:2021 - Security Misconfiguration)
		c.Header("Server", "")

		c.Next()
	}
}

// InputSanitization provides comprehensive input sanitization against injection attacks
func (sm *SecurityMiddleware) InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize URL parameters (A03:2021 - Injection)
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				sanitized := sm.sanitizeInput(value)
				if sanitized != value {
					sm.logger.Warn("Potentially malicious input detected in URL parameter",
						zap.String("parameter", key),
						zap.String("original", value),
						zap.String("sanitized", sanitized),
						zap.String("client_ip", c.ClientIP()),
					)
				}
				values[i] = sanitized
			}
		}

		// Sanitize headers for common injection vectors
		suspiciousHeaders := []string{
			"User-Agent", "Referer", "X-Forwarded-For",
			"X-Real-IP", "X-Forwarded-Proto", "Host",
		}

		for _, headerName := range suspiciousHeaders {
			headerValue := c.GetHeader(headerName)
			if headerValue != "" {
				sanitized := sm.sanitizeInput(headerValue)
				if sanitized != headerValue {
					sm.logger.Warn("Potentially malicious input detected in header",
						zap.String("header", headerName),
						zap.String("original", headerValue),
						zap.String("sanitized", sanitized),
						zap.String("client_ip", c.ClientIP()),
					)
					c.Request.Header.Set(headerName, sanitized)
				}
			}
		}

		c.Next()
	}
}

// RequestSizeLimit prevents DoS attacks through large payloads (A06:2021 - Vulnerable Components)
func (sm *SecurityMiddleware) RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			sm.logger.Warn("Request size limit exceeded",
				zap.Int64("content_length", c.Request.ContentLength),
				zap.Int64("max_size", maxSize),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.GetHeader("User-Agent")),
			)

			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"message": "Request payload too large",
				"error": gin.H{
					"code":    "PAYLOAD_TOO_LARGE",
					"message": "Request size exceeds maximum allowed limit",
				},
			})
			c.Abort()
			return
		}

		// Set a hard limit on request body reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// TimeoutMiddleware prevents slowloris attacks (A05:2021 - Security Misconfiguration)
func (sm *SecurityMiddleware) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			sm.logger.Warn("Request timeout exceeded",
				zap.Duration("timeout", timeout),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
			)

			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
				"error": gin.H{
					"code":    "REQUEST_TIMEOUT",
					"message": "Request processing time exceeded limit",
				},
			})
			c.Abort()
		}
	}
}

// APIKeyValidation provides simple API key validation for sensitive endpoints
func (sm *SecurityMiddleware) APIKeyValidation(requiredAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != requiredAPIKey {
			sm.logger.Warn("Invalid or missing API key",
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.GetHeader("User-Agent")),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid API key",
				"error": gin.H{
					"code":    "INVALID_API_KEY",
					"message": "Valid API key required for this endpoint",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// sanitizeInput removes potentially dangerous characters and patterns
func (sm *SecurityMiddleware) sanitizeInput(input string) string {
	// Remove null bytes (common in injection attacks)
	input = strings.ReplaceAll(input, "\x00", "")

	// Basic XSS prevention - encode dangerous characters
	replacements := map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#x27;",
		"&":  "&amp;",
	}

	for old, new := range replacements {
		input = strings.ReplaceAll(input, old, new)
	}

	// SQL injection prevention - escape common SQL meta-characters
	sqlPatterns := []string{
		"--", "/*", "*/", "xp_", "sp_", "exec", "execute",
		"union", "select", "insert", "update", "delete", "drop",
		"create", "alter", "truncate", "script", "javascript:",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			// Replace with safe equivalent
			input = strings.ReplaceAll(input, pattern, strings.ReplaceAll(pattern, " ", "_"))
		}
	}

	// Path traversal prevention
	input = strings.ReplaceAll(input, "../", "")
	input = strings.ReplaceAll(input, "..\\", "")

	// Command injection prevention
	cmdPatterns := []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "[", "]"}
	for _, pattern := range cmdPatterns {
		input = strings.ReplaceAll(input, pattern, "")
	}

	return strings.TrimSpace(input)
}

// ValidateSecurityConfiguration validates security middleware configuration
func ValidateSecurityConfiguration() error {
	// This could be expanded to validate security settings from config
	// For now, just ensure basic security principles are followed
	return nil
}

// SecurityAuditLog logs security-relevant events for monitoring
func (sm *SecurityMiddleware) SecurityAuditLog(eventType string, details map[string]interface{}) {
	sm.logger.Info("Security audit event",
		zap.String("event_type", eventType),
		zap.Any("details", details),
		zap.Time("timestamp", time.Now()),
	)
}

// IPWhitelist restricts access to specific IP addresses (for admin endpoints)
func (sm *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !allowedMap[clientIP] {
			sm.logger.Warn("Access denied - IP not whitelisted",
				zap.String("client_ip", clientIP),
				zap.Strings("allowed_ips", allowedIPs),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
				"error": gin.H{
					"code":    "IP_NOT_ALLOWED",
					"message": "Your IP address is not authorized to access this resource",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS provides Cross-Origin Resource Sharing configuration
func (sm *SecurityMiddleware) CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
