// Package auth provides JWT token generation and validation utilities
package auth

import (
	"context"
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

// BlacklistStore defines the interface for JWT blacklist storage
type BlacklistStore interface {
	// AddToken adds a token to the blacklist with expiration time
	AddToken(tokenString string, expiresAt time.Time) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(tokenString string) (bool, error)

	// CleanupExpired removes expired tokens from blacklist
	CleanupExpired() error
}

// JWTService provides JWT operations with configuration injection and blacklist support
type JWTService struct {
	config         *config.Config
	blacklistStore BlacklistStore
}

// NewJWTService creates a new JWT service with configuration
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

// NewJWTServiceWithBlacklist creates a new JWT service with configuration and blacklist
func NewJWTServiceWithBlacklist(cfg *config.Config, blacklistStore BlacklistStore) *JWTService {
	return &JWTService{
		config:         cfg,
		blacklistStore: blacklistStore,
	}
}

// GenerateToken generates a JWT token using the service's configuration
func (s *JWTService) GenerateToken(userID uuid.UUID) (string, time.Time, error) {
	return GenerateToken(userID, s.config.JWTSecret)
}

// ValidateToken validates a JWT token using the service's configuration and checks blacklist
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// First check if token is blacklisted
	if s.blacklistStore != nil {
		isBlacklisted, err := s.blacklistStore.IsBlacklisted(tokenString)
		if err != nil {
			return nil, fmt.Errorf("blacklist check failed: %w", err)
		}
		if isBlacklisted {
			return nil, errors.New("token has been blacklisted")
		}
	}

	return ValidateToken(tokenString, s.config.JWTSecret)
}

// BlacklistToken adds a token to the blacklist
func (s *JWTService) BlacklistToken(ctx context.Context, tokenString string) error {
	if s.blacklistStore == nil {
		return errors.New("blacklist store not configured")
	}

	// Parse token to get expiration time
	claims, err := ValidateToken(tokenString, s.config.JWTSecret)
	if err != nil {
		// Even if token is invalid, we still want to blacklist it
		// Use a default expiration time
		expiresAt := time.Now().Add(24 * time.Hour)
		return s.blacklistStore.AddToken(tokenString, expiresAt)
	}

	return s.blacklistStore.AddToken(tokenString, claims.ExpiresAt.Time)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *JWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	if s.blacklistStore == nil {
		return false, nil // No blacklist store means no tokens are blacklisted
	}

	return s.blacklistStore.IsBlacklisted(tokenString)
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (s *JWTService) CleanupExpiredTokens() error {
	if s.blacklistStore == nil {
		return nil // No blacklist store means nothing to clean up
	}

	return s.blacklistStore.CleanupExpired()
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
