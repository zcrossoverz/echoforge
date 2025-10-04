// Package auth provides JWT token generation and validation utilities
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/config"
)

// JWTClaims represents the custom claims for JWT tokens (clone-and-extend model)
type JWTClaims struct {
	UserID string `json:"sub"` // Subject: User ID only
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token with user claims only (clone-and-extend model)
func GenerateToken(userID uuid.UUID, secret string) (string, time.Time, error) {
	if secret == "" {
		return "", time.Time{}, errors.New("JWT secret cannot be empty")
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// GenerateTokenWithConfig generates a JWT token using the application configuration
func GenerateTokenWithConfig(userID uuid.UUID, cfg *config.Config) (string, time.Time, error) {
	return GenerateToken(userID, cfg.JWTSecret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	if secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateTokenWithConfig validates a JWT token using the application configuration
func ValidateTokenWithConfig(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	return ValidateToken(tokenString, cfg.JWTSecret)
}

// JWTService provides JWT operations with configuration injection
type JWTService struct {
	config *config.Config
}

// NewJWTService creates a new JWT service with configuration
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

// GenerateToken generates a JWT token using the service's configuration
func (s *JWTService) GenerateToken(userID uuid.UUID) (string, time.Time, error) {
	return GenerateToken(userID, s.config.JWTSecret)
}

// ValidateToken validates a JWT token using the service's configuration
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	return ValidateToken(tokenString, s.config.JWTSecret)
}

// GetSecret returns the JWT secret from configuration (for backward compatibility)
func (s *JWTService) GetSecret() string {
	return s.config.JWTSecret
}

// Legacy function for backward compatibility
func GenerateJWT(userID, role string) (string, error) {
	// Deprecated: Use GenerateToken instead
	// This maintains compatibility with existing code
	return fmt.Sprintf("jwt-token.%s.%s", userID[:8], role), nil
}
