package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest        ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeMethodNotAllowed  ErrorCode = "METHOD_NOT_ALLOWED"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodePayloadTooLarge   ErrorCode = "PAYLOAD_TOO_LARGE"

	// Server errors (5xx)
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            ErrorCode = "TIMEOUT"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService    ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// APIError represents a standardized API error response
type APIError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Internal   error                  `json:"-"` // Internal error for logging, not exposed
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Error   APIError `json:"error"`
}

// ErrorHandler provides centralized error handling with information disclosure protection
type ErrorHandler struct {
	logger      *zap.Logger
	development bool // If true, includes more details in error responses
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger, development bool) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger,
		development: development,
	}
}

// HandleError processes errors and returns appropriate responses
func (eh *ErrorHandler) HandleError(c *gin.Context, err error) {
	var apiErr *APIError

	// Convert error to APIError
	if !errors.As(err, &apiErr) {
		// Unknown error - sanitize for security
		apiErr = eh.sanitizeError(err)
	}

	// Log the error with full context
	eh.logError(c, apiErr)

	// Prepare response (sanitized for client)
	response := ErrorResponse{
		Success: false,
		Message: eh.getPublicMessage(apiErr),
		Error: APIError{
			Code:       apiErr.Code,
			Message:    eh.getPublicMessage(apiErr),
			Details:    eh.getPublicDetails(apiErr),
			StatusCode: apiErr.StatusCode,
		},
	}

	c.JSON(apiErr.StatusCode, response)
}

// ErrorMiddleware provides Gin middleware for centralized error handling
func (eh *ErrorHandler) ErrorMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch x := recovered.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = fmt.Errorf("unknown panic: %v", x)
		}

		// Log panic with stack trace
		eh.logPanic(c, err, recovered)

		// Return generic internal server error
		apiErr := &APIError{
			Code:       ErrCodeInternalError,
			Message:    "An unexpected error occurred",
			StatusCode: http.StatusInternalServerError,
			Internal:   err,
		}

		eh.HandleError(c, apiErr)
	})
}

// Common error constructors

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string, details map[string]interface{}) *APIError {
	return &APIError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *APIError {
	if message == "" {
		message = "Authentication required"
	}
	return &APIError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *APIError {
	if message == "" {
		message = "Access denied"
	}
	return &APIError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *APIError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return &APIError{
		Code:       ErrCodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details map[string]interface{}) *APIError {
	if message == "" {
		message = "Validation failed"
	}
	return &APIError{
		Code:       ErrCodeValidationFailed,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(internalErr error) *APIError {
	return &APIError{
		Code:       ErrCodeInternalError,
		Message:    "An internal error occurred",
		StatusCode: http.StatusInternalServerError,
		Internal:   internalErr,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(internalErr error) *APIError {
	return &APIError{
		Code:       ErrCodeDatabaseError,
		Message:    "Database operation failed",
		StatusCode: http.StatusInternalServerError,
		Internal:   internalErr,
	}
}

// sanitizeError converts unknown errors to safe APIErrors
func (eh *ErrorHandler) sanitizeError(err error) *APIError {
	// Check for common error types and map to appropriate API errors
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "database"):
		return NewDatabaseError(err)
	case strings.Contains(errStr, "timeout"):
		return &APIError{
			Code:       ErrCodeTimeout,
			Message:    "Request timeout",
			StatusCode: http.StatusRequestTimeout,
			Internal:   err,
		}
	case strings.Contains(errStr, "connection"):
		return &APIError{
			Code:       ErrCodeServiceUnavailable,
			Message:    "Service temporarily unavailable",
			StatusCode: http.StatusServiceUnavailable,
			Internal:   err,
		}
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return NewBadRequestError("Invalid request", nil)
	case strings.Contains(errStr, "not found"):
		return NewNotFoundError("")
	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication"):
		return NewUnauthorizedError("")
	case strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "permission"):
		return NewForbiddenError("")
	default:
		// Unknown error - return generic internal error
		return NewInternalError(err)
	}
}

// getPublicMessage returns a safe message for client consumption
func (eh *ErrorHandler) getPublicMessage(apiErr *APIError) string {
	if eh.development {
		// In development, return the actual message
		return apiErr.Message
	}

	// In production, return generic messages for security
	switch apiErr.Code {
	case ErrCodeBadRequest, ErrCodeValidationFailed:
		return "Invalid request data"
	case ErrCodeUnauthorized:
		return "Authentication required"
	case ErrCodeForbidden:
		return "Access denied"
	case ErrCodeNotFound:
		return "Resource not found"
	case ErrCodeMethodNotAllowed:
		return "Method not allowed"
	case ErrCodeConflict:
		return "Request conflicts with current state"
	case ErrCodeRateLimitExceeded:
		return "Too many requests"
	case ErrCodePayloadTooLarge:
		return "Request payload too large"
	case ErrCodeTimeout:
		return "Request timeout"
	case ErrCodeServiceUnavailable:
		return "Service temporarily unavailable"
	default:
		return "An error occurred while processing your request"
	}
}

// getPublicDetails returns safe details for client consumption
func (eh *ErrorHandler) getPublicDetails(apiErr *APIError) map[string]interface{} {
	if !eh.development {
		// In production, don't expose internal details
		return nil
	}

	// In development, return sanitized details
	if apiErr.Details == nil {
		return nil
	}

	publicDetails := make(map[string]interface{})
	for key, value := range apiErr.Details {
		// Only include safe fields
		switch key {
		case "field", "validation_errors", "required_fields", "allowed_values":
			publicDetails[key] = value
		default:
			// Skip potentially sensitive fields
		}
	}

	return publicDetails
}

// logError logs error details for debugging and monitoring
func (eh *ErrorHandler) logError(c *gin.Context, apiErr *APIError) {
	fields := []zap.Field{
		zap.String("error_code", string(apiErr.Code)),
		zap.String("public_message", apiErr.Message),
		zap.Int("status_code", apiErr.StatusCode),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	}

	// Add internal error details if available
	if apiErr.Internal != nil {
		fields = append(fields, zap.Error(apiErr.Internal))
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		fields = append(fields, zap.Any("user_id", userID))
	}

	// Add request details
	if apiErr.Details != nil {
		fields = append(fields, zap.Any("error_details", apiErr.Details))
	}

	// Log at appropriate level
	switch apiErr.StatusCode {
	case http.StatusInternalServerError:
		eh.logger.Error("API error", fields...)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		eh.logger.Warn("API error", fields...)
	default:
		eh.logger.Info("API error", fields...)
	}
}

// logPanic logs panic details with stack trace
func (eh *ErrorHandler) logPanic(c *gin.Context, err error, recovered interface{}) {
	// Get stack trace
	stack := make([]byte, 4096)
	stack = stack[:runtime.Stack(stack, false)]

	eh.logger.Error("Panic recovered",
		zap.Any("panic", recovered),
		zap.Error(err),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("stack_trace", string(stack)),
	)
}

// ValidationErrorFromMap creates a validation error from a map of field errors
func ValidationErrorFromMap(fieldErrors map[string]string) *APIError {
	details := make(map[string]interface{})
	details["validation_errors"] = fieldErrors

	return NewValidationError("Validation failed", details)
}

// IsClientError checks if the error is a client error (4xx)
func IsClientError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
	}
	return false
}

// IsServerError checks if the error is a server error (5xx)
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// GetStatusCode extracts HTTP status code from error
func GetStatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return http.StatusInternalServerError
}
