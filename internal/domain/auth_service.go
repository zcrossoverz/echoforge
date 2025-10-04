package domain

import (
	"context"
	"errors"
	"time"
)

// Authentication domain errors
var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrTokenExpired          = errors.New("token has expired")
	ErrTokenInvalid          = errors.New("token is invalid")
	ErrTokenBlacklisted      = errors.New("token has been blacklisted")
	ErrAuthorizationRequired = errors.New("authorization required")
)

// AuthToken represents a JWT authentication token
type AuthToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Subject   string    `json:"subject"` // User ID
}

// AuthService defines the contract for authentication operations
// This interface abstracts JWT operations from the business logic
type AuthService interface {
	// GenerateToken creates a new JWT token for the given user
	// Token should be valid for 24 hours as per specification
	GenerateToken(ctx context.Context, user *User) (*AuthToken, error)

	// ValidateToken verifies a JWT token and returns the user ID
	// Returns ErrTokenInvalid, ErrTokenExpired, or ErrTokenBlacklisted as appropriate
	ValidateToken(ctx context.Context, tokenString string) (userID string, err error)

	// RefreshToken generates a new token from a valid existing token
	// Useful for extending user sessions without re-authentication
	RefreshToken(ctx context.Context, tokenString string) (*AuthToken, error)

	// BlacklistToken adds a token to the blacklist (for logout functionality)
	// Blacklisted tokens should be rejected by ValidateToken
	BlacklistToken(ctx context.Context, tokenString string) error

	// IsTokenBlacklisted checks if a token has been blacklisted
	IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error)

	// CleanupExpiredTokens removes expired tokens from blacklist (maintenance operation)
	// Should be called periodically to prevent blacklist from growing indefinitely
	CleanupExpiredTokens(ctx context.Context) error
}

// PasswordService defines the contract for password operations
// Abstracts bcrypt operations from business logic
type PasswordService interface {
	// HashPassword generates a bcrypt hash from a plain text password
	// Should use cost factor 12 as per specification
	HashPassword(password string) (string, error)

	// VerifyPassword compares a plain text password with a bcrypt hash
	// Returns true if password matches the hash
	VerifyPassword(password, hash string) bool

	// ValidatePassword checks if a password meets security requirements
	// Returns appropriate validation errors
	ValidatePassword(password string) error
}
