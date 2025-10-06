// Package contracts defines the updated authentication interfaces for clone-and-extend model
package contracts

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the updated custom claims for JWT tokens (site_id removed)
type JWTClaims struct {
	UserID string `json:"sub"` // Subject: User ID only
	jwt.RegisteredClaims
}

// TokenGenerator defines the interface for JWT token generation
type TokenGenerator interface {
	// GenerateToken generates a JWT token with user claims only (site_id removed)
	GenerateToken(userID uuid.UUID, secret string) (string, time.Time, error)
}

// TokenValidator defines the interface for JWT token validation
type TokenValidator interface {
	// ValidateToken validates a JWT token and returns the claims
	ValidateToken(tokenString, secret string) (*JWTClaims, error)
}

// AuthService combines token generation and validation
type AuthService interface {
	TokenGenerator
	TokenValidator
}
