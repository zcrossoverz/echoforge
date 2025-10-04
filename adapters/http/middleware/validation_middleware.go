package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware provides request validation capabilities
type ValidationMiddleware struct {
	validator *validator.Validate
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	validate := validator.New()

	// Register custom validation functions
	registerCustomValidations(validate)

	return &ValidationMiddleware{
		validator: validate,
	}
}

// ValidationResponse represents validation error response
type ValidationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// ValidateJSON validates JSON request bodies against a struct type
func (vm *ValidationMiddleware) ValidateJSON(structType interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Success: false,
				Message: "Failed to read request body",
				Error: struct {
					Code    string                 `json:"code"`
					Message string                 `json:"message"`
					Details map[string]interface{} `json:"details,omitempty"`
				}{
					Code:    "BODY_READ_ERROR",
					Message: "Unable to read request body",
					Details: map[string]interface{}{"error": err.Error()},
				},
			})
			c.Abort()
			return
		}

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Skip validation for empty body if not required
		if len(body) == 0 {
			c.Next()
			return
		}

		// Create new instance of the struct type
		structValue := reflect.New(reflect.TypeOf(structType).Elem()).Interface()

		// Unmarshal JSON
		if err := json.Unmarshal(body, structValue); err != nil {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Success: false,
				Message: "Invalid JSON format",
				Error: struct {
					Code    string                 `json:"code"`
					Message string                 `json:"message"`
					Details map[string]interface{} `json:"details,omitempty"`
				}{
					Code:    "INVALID_JSON",
					Message: "Request body contains invalid JSON",
					Details: map[string]interface{}{"error": err.Error()},
				},
			})
			c.Abort()
			return
		}

		// Validate struct
		if err := vm.validator.Struct(structValue); err != nil {
			validationErrors := vm.formatValidationErrors(err)

			c.JSON(http.StatusBadRequest, ValidationResponse{
				Success: false,
				Message: "Validation failed",
				Error: struct {
					Code    string                 `json:"code"`
					Message string                 `json:"message"`
					Details map[string]interface{} `json:"details,omitempty"`
				}{
					Code:    "VALIDATION_ERROR",
					Message: "Request validation failed",
					Details: validationErrors,
				},
			})
			c.Abort()
			return
		}

		// Store validated struct in context
		c.Set("validated_data", structValue)
		c.Next()
	}
}

// SanitizeInput sanitizes input to prevent XSS and injection attacks
func (vm *ValidationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read and sanitize request body if it's JSON
		if c.ContentType() == "application/json" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}

			// Parse JSON
			var jsonData map[string]interface{}
			if err := json.Unmarshal(body, &jsonData); err != nil {
				// Not valid JSON, restore body and continue
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				c.Next()
				return
			}

			// Sanitize JSON data
			sanitizedData := vm.sanitizeJSONData(jsonData)

			// Marshal back to JSON
			sanitizedBody, err := json.Marshal(sanitizedData)
			if err != nil {
				// Error marshaling, use original body
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				c.Next()
				return
			}

			// Replace request body with sanitized version
			c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
			c.Request.ContentLength = int64(len(sanitizedBody))
		}

		c.Next()
	}
}

// ValidateContentType ensures the request has the expected content type
func (vm *ValidationMiddleware) ValidateContentType(expectedType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.ContentType()

		if contentType != expectedType {
			c.JSON(http.StatusUnsupportedMediaType, ValidationResponse{
				Success: false,
				Message: "Unsupported content type",
				Error: struct {
					Code    string                 `json:"code"`
					Message string                 `json:"message"`
					Details map[string]interface{} `json:"details,omitempty"`
				}{
					Code:    "UNSUPPORTED_CONTENT_TYPE",
					Message: fmt.Sprintf("Expected content type '%s', got '%s'", expectedType, contentType),
					Details: map[string]interface{}{
						"expected": expectedType,
						"received": contentType,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetValidatedData retrieves validated data from the context
func GetValidatedData(c *gin.Context) (interface{}, bool) {
	if data, exists := c.Get("validated_data"); exists {
		return data, true
	}
	return nil, false
}

// Helper functions

// formatValidationErrors converts validator errors to a structured format
func (vm *ValidationMiddleware) formatValidationErrors(err error) map[string]interface{} {
	errors := make(map[string]interface{})

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationError := range validationErrors {
			fieldName := validationError.Field()
			errors[fieldName] = vm.getErrorMessage(validationError)
		}
	}

	return errors
}

// getErrorMessage returns a user-friendly error message for validation errors
func (vm *ValidationMiddleware) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", err.Field(), err.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", err.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", err.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "password_strength":
		return "Password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"
	default:
		return fmt.Sprintf("%s validation failed", err.Field())
	}
}

// sanitizeJSONData recursively sanitizes JSON data
func (vm *ValidationMiddleware) sanitizeJSONData(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[key] = vm.sanitizeString(v)
		case map[string]interface{}:
			sanitized[key] = vm.sanitizeJSONData(v)
		case []interface{}:
			sanitized[key] = vm.sanitizeArray(v)
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// sanitizeArray sanitizes array elements
func (vm *ValidationMiddleware) sanitizeArray(data []interface{}) []interface{} {
	sanitized := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[i] = vm.sanitizeString(v)
		case map[string]interface{}:
			sanitized[i] = vm.sanitizeJSONData(v)
		case []interface{}:
			sanitized[i] = vm.sanitizeArray(v)
		default:
			sanitized[i] = value
		}
	}

	return sanitized
}

// sanitizeString removes potentially dangerous characters and scripts
func (vm *ValidationMiddleware) sanitizeString(input string) string {
	// Remove HTML tags
	input = vm.stripHTMLTags(input)

	// Remove SQL injection patterns (basic protection)
	input = vm.removeSQLPatterns(input)

	// Remove JavaScript patterns
	input = vm.removeJSPatterns(input)

	return strings.TrimSpace(input)
}

// stripHTMLTags removes HTML tags from input
func (vm *ValidationMiddleware) stripHTMLTags(input string) string {
	// Simple HTML tag removal (more sophisticated libraries available for production)
	input = strings.ReplaceAll(input, "<script", "&lt;script")
	input = strings.ReplaceAll(input, "</script>", "&lt;/script&gt;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	return input
}

// removeSQLPatterns removes common SQL injection patterns
func (vm *ValidationMiddleware) removeSQLPatterns(input string) string {
	lowerInput := strings.ToLower(input)

	// Common SQL injection patterns
	patterns := []string{
		"drop table", "delete from", "insert into", "update set",
		"union select", "'; drop", "--", "/*", "*/",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerInput, pattern) {
			// Replace with safe equivalent or remove
			input = strings.ReplaceAll(input, pattern, "")
		}
	}

	return input
}

// removeJSPatterns removes JavaScript patterns
func (vm *ValidationMiddleware) removeJSPatterns(input string) string {
	lowerInput := strings.ToLower(input)

	// Common JavaScript patterns
	patterns := []string{
		"javascript:", "onclick=", "onerror=", "onload=",
		"eval(", "alert(", "confirm(", "prompt(",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerInput, pattern) {
			input = strings.ReplaceAll(input, pattern, "")
		}
	}

	return input
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations(validate *validator.Validate) {
	// Password strength validation
	validate.RegisterValidation("password_strength", validatePasswordStrength)
}

// validatePasswordStrength validates password meets strength requirements
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}
