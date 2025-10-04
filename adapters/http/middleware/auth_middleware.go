package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	authUseCase *usecase.UserAuthenticationUseCase
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authUseCase *usecase.UserAuthenticationUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
	}
}

// AuthResponse represents the authentication error response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	} `json:"error"`
}

// RequireAuth middleware function that requires valid JWT authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Authorization required",
				Error: struct {
					Code    string `json:"code"`
					Message string `json:"message"`
					Details string `json:"details,omitempty"`
				}{
					Code:    "MISSING_TOKEN",
					Message: "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		token, err := extractBearerToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Invalid authorization header format",
				Error: struct {
					Code    string `json:"code"`
					Message string `json:"message"`
					Details string `json:"details,omitempty"`
				}{
					Code:    "INVALID_TOKEN_FORMAT",
					Message: "Authorization header must be in format 'Bearer <token>'",
					Details: err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Authenticate with token
		userDTO, err := m.authUseCase.AuthenticateWithToken(c.Request.Context(), token)
		if err != nil {
			statusCode, errorCode, message := mapAuthError(err)

			c.JSON(statusCode, AuthResponse{
				Success: false,
				Message: message,
				Error: struct {
					Code    string `json:"code"`
					Message string `json:"message"`
					Details string `json:"details,omitempty"`
				}{
					Code:    errorCode,
					Message: message,
					Details: err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Set user information in context for downstream handlers
		c.Set("user", userDTO)
		c.Set("user_id", userDTO.ID)
		c.Set("token", token)

		// Continue to next handler
		c.Next()
	}
}

// OptionalAuth middleware function that optionally authenticates but doesn't block unauthenticated requests
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Parse Bearer token
		token, err := extractBearerToken(authHeader)
		if err != nil {
			// Invalid token format, continue without authentication
			c.Next()
			return
		}

		// Attempt to authenticate with token
		userDTO, err := m.authUseCase.AuthenticateWithToken(c.Request.Context(), token)
		if err != nil {
			// Authentication failed, continue without authentication
			c.Next()
			return
		}

		// Set user information in context for downstream handlers
		c.Set("user", userDTO)
		c.Set("user_id", userDTO.ID)
		c.Set("token", token)

		// Continue to next handler
		c.Next()
	}
}

// GetCurrentUser retrieves the current authenticated user from the Gin context
func GetCurrentUser(c *gin.Context) (*usecase.UserDTO, bool) {
	if user, exists := c.Get("user"); exists {
		if userDTO, ok := user.(*usecase.UserDTO); ok {
			return userDTO, true
		}
	}
	return nil, false
}

// GetCurrentUserID retrieves the current authenticated user ID from the Gin context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// GetCurrentToken retrieves the current JWT token from the Gin context
func GetCurrentToken(c *gin.Context) (string, bool) {
	if token, exists := c.Get("token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr, true
		}
	}
	return "", false
}

// Helper functions

// extractBearerToken extracts the JWT token from the Authorization header
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", domain.ErrAuthorizationRequired
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", domain.ErrTokenInvalid
	}

	if parts[1] == "" {
		return "", domain.ErrTokenInvalid
	}

	return parts[1], nil
}

// mapAuthError maps authentication errors to HTTP status codes and error codes
func mapAuthError(err error) (int, string, string) {
	switch {
	case err == domain.ErrTokenInvalid:
		return http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid"
	case err == domain.ErrTokenExpired:
		return http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired"
	case err == domain.ErrTokenBlacklisted:
		return http.StatusUnauthorized, "TOKEN_BLACKLISTED", "Token has been blacklisted"
	case err == domain.ErrAuthorizationRequired:
		return http.StatusUnauthorized, "AUTHORIZATION_REQUIRED", "Authorization is required"
	case strings.Contains(err.Error(), "user not found"):
		return http.StatusUnauthorized, "USER_NOT_FOUND", "User associated with token not found"
	default:
		return http.StatusInternalServerError, "AUTHENTICATION_ERROR", "Authentication failed due to server error"
	}
}
