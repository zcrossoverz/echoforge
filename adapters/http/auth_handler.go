package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRegistrationUC   *usecase.UserRegistrationUseCase
	userAuthenticationUC *usecase.UserAuthenticationUseCase
	userLogoutUC         *usecase.UserLogoutUseCase
	getUserProfileUC     *usecase.GetUserProfileUseCase
	validator            *validator.Validate
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	userRegistrationUC *usecase.UserRegistrationUseCase,
	userAuthenticationUC *usecase.UserAuthenticationUseCase,
	userLogoutUC *usecase.UserLogoutUseCase,
	getUserProfileUC *usecase.GetUserProfileUseCase,
) *AuthHandler {
	return &AuthHandler{
		userRegistrationUC:   userRegistrationUC,
		userAuthenticationUC: userAuthenticationUC,
		userLogoutUC:         userLogoutUC,
		getUserProfileUC:     getUserProfileUC,
		validator:            validator.New(),
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// RegisterResponse represents the registration response payload
type RegisterResponse struct {
	Success bool                          `json:"success"`
	Message string                        `json:"message"`
	Data    *usecase.RegisterUserResponse `json:"data,omitempty"`
	Error   *ErrorResponse                `json:"error,omitempty"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Message: "Invalid request format",
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Request body is not valid JSON",
				Details: err.Error(),
			},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Message: "Validation failed",
			Error: &ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
				Details: getValidationErrorMessage(err),
			},
		})
		return
	}

	// Convert to use case request
	ucRequest := &usecase.RegisterUserRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Execute use case
	result, err := h.userRegistrationUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		// Handle different error types
		statusCode, errorCode, message := mapRegistrationError(err)

		c.JSON(statusCode, RegisterResponse{
			Success: false,
			Message: message,
			Error: &ErrorResponse{
				Code:    errorCode,
				Message: message,
				Details: err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    result,
	})
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    *usecase.LoginResponse `json:"data,omitempty"`
	Error   *ErrorResponse         `json:"error,omitempty"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request format",
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Request body is not valid JSON",
				Details: err.Error(),
			},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Validation failed",
			Error: &ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
				Details: getValidationErrorMessage(err),
			},
		})
		return
	}

	// Convert to use case request
	ucRequest := &usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Execute use case
	result, err := h.userAuthenticationUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		// Handle different error types
		statusCode, errorCode, message := mapAuthenticationError(err)

		c.JSON(statusCode, LoginResponse{
			Success: false,
			Message: message,
			Error: &ErrorResponse{
				Code:    errorCode,
				Message: message,
				Details: err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "Login successful",
		Data:    result,
	})
}

// LogoutResponse represents the logout response payload
type LogoutResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract JWT token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, LogoutResponse{
			Success: false,
			Message: "Authorization required",
			Error: &ErrorResponse{
				Code:    "MISSING_TOKEN",
				Message: "Authorization header is required",
			},
		})
		return
	}

	// Parse Bearer token
	token, err := extractBearerToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LogoutResponse{
			Success: false,
			Message: "Invalid authorization header format",
			Error: &ErrorResponse{
				Code:    "INVALID_TOKEN_FORMAT",
				Message: "Authorization header must be in format 'Bearer <token>'",
				Details: err.Error(),
			},
		})
		return
	}

	// Execute logout use case
	result, err := h.userLogoutUC.ExecuteWithToken(c.Request.Context(), token)
	if err != nil {
		// Handle different error types
		statusCode, errorCode, message := mapLogoutError(err)

		c.JSON(statusCode, LogoutResponse{
			Success: false,
			Message: message,
			Error: &ErrorResponse{
				Code:    errorCode,
				Message: message,
				Details: err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, LogoutResponse{
		Success: result.Success,
		Message: result.Message,
	})
}

// ProfileResponse represents the profile response payload
type ProfileResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    *usecase.GetProfileResponse `json:"data,omitempty"`
	Error   *ErrorResponse              `json:"error,omitempty"`
}

// GetProfile handles GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Extract JWT token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ProfileResponse{
			Success: false,
			Message: "Authorization required",
			Error: &ErrorResponse{
				Code:    "MISSING_TOKEN",
				Message: "Authorization header is required",
			},
		})
		return
	}

	// Parse Bearer token
	token, err := extractBearerToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ProfileResponse{
			Success: false,
			Message: "Invalid authorization header format",
			Error: &ErrorResponse{
				Code:    "INVALID_TOKEN_FORMAT",
				Message: "Authorization header must be in format 'Bearer <token>'",
				Details: err.Error(),
			},
		})
		return
	}

	// Execute get profile use case
	result, err := h.getUserProfileUC.ExecuteWithToken(c.Request.Context(), token)
	if err != nil {
		// Handle different error types
		statusCode, errorCode, message := mapProfileError(err)

		c.JSON(statusCode, ProfileResponse{
			Success: false,
			Message: message,
			Error: &ErrorResponse{
				Code:    errorCode,
				Message: message,
				Details: err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, ProfileResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    result,
	})
}

// Helper functions

// extractBearerToken extracts the JWT token from the Authorization header
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header must be in format 'Bearer <token>'")
	}

	return parts[1], nil
}

// getValidationErrorMessage converts validation errors to user-friendly messages
func getValidationErrorMessage(err error) string {
	var validationErrors []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrs {
			switch validationErr.Tag() {
			case "required":
				validationErrors = append(validationErrors, validationErr.Field()+" is required")
			case "email":
				validationErrors = append(validationErrors, validationErr.Field()+" must be a valid email address")
			case "min":
				validationErrors = append(validationErrors, validationErr.Field()+" must be at least "+validationErr.Param()+" characters long")
			case "max":
				validationErrors = append(validationErrors, validationErr.Field()+" must not exceed "+validationErr.Param()+" characters")
			default:
				validationErrors = append(validationErrors, validationErr.Field()+" validation failed")
			}
		}
		return strings.Join(validationErrors, ", ")
	}

	return err.Error()
}

// mapRegistrationError maps registration errors to HTTP status codes and error messages
func mapRegistrationError(err error) (int, string, string) {
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		return http.StatusConflict, "USER_ALREADY_EXISTS", "A user with this email already exists"
	}
	if errors.Is(err, domain.ErrInvalidEmail) {
		return http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format"
	}
	if errors.Is(err, domain.ErrEmailTooLong) {
		return http.StatusBadRequest, "EMAIL_TOO_LONG", "Email address is too long"
	}

	// Password validation errors
	if strings.Contains(err.Error(), "password") {
		return http.StatusBadRequest, "INVALID_PASSWORD", "Password does not meet security requirements"
	}

	// Generic server error
	return http.StatusInternalServerError, "REGISTRATION_FAILED", "Registration failed due to server error"
}

// mapAuthenticationError maps authentication errors to HTTP status codes and error messages
func mapAuthenticationError(err error) (int, string, string) {
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password"
	}
	if errors.Is(err, domain.ErrInvalidEmail) {
		return http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format"
	}

	// Generic server error
	return http.StatusInternalServerError, "AUTHENTICATION_FAILED", "Authentication failed due to server error"
}

// mapLogoutError maps logout errors to HTTP status codes and error messages
func mapLogoutError(err error) (int, string, string) {
	if errors.Is(err, domain.ErrTokenInvalid) {
		return http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid"
	}
	if errors.Is(err, domain.ErrTokenExpired) {
		return http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired"
	}
	if errors.Is(err, domain.ErrTokenBlacklisted) {
		return http.StatusUnauthorized, "TOKEN_BLACKLISTED", "Token has been blacklisted"
	}
	if errors.Is(err, domain.ErrAuthorizationRequired) {
		return http.StatusUnauthorized, "AUTHORIZATION_REQUIRED", "Authorization is required"
	}

	// Generic server error
	return http.StatusInternalServerError, "LOGOUT_FAILED", "Logout failed due to server error"
}

// mapProfileError maps profile retrieval errors to HTTP status codes and error messages
func mapProfileError(err error) (int, string, string) {
	if errors.Is(err, domain.ErrTokenInvalid) {
		return http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid"
	}
	if errors.Is(err, domain.ErrTokenExpired) {
		return http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired"
	}
	if errors.Is(err, domain.ErrTokenBlacklisted) {
		return http.StatusUnauthorized, "TOKEN_BLACKLISTED", "Token has been blacklisted"
	}
	if errors.Is(err, domain.ErrAuthorizationRequired) {
		return http.StatusUnauthorized, "AUTHORIZATION_REQUIRED", "Authorization is required"
	}
	if strings.Contains(err.Error(), "user not found") {
		return http.StatusNotFound, "USER_NOT_FOUND", "User not found"
	}

	// Generic server error
	return http.StatusInternalServerError, "PROFILE_RETRIEVAL_FAILED", "Profile retrieval failed due to server error"
}
