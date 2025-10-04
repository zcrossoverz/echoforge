// Package contracts defines the updated domain interfaces for clone-and-extend model
package contracts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents the updated user entity without site_id
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository defines the updated persistence contract for User entities
type UserRepository interface {
	// Create persists a new user entity
	Create(ctx context.Context, user *User) error

	// FindByEmail retrieves a user by email (site_id parameter removed)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

// Validation interface for user entity
type UserValidator interface {
	// Validate performs business rule validation
	Validate() error

	// IsValid returns true if entity passes all validation rules
	IsValid() bool
}

// Domain errors (unchanged from current implementation)
var (
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrEmailTooLong         = errors.New("email exceeds maximum length")
	ErrPasswordHashTooShort = errors.New("password hash too short")
	ErrRequiredField        = errors.New("required field is empty")
	ErrUserAlreadyExists    = errors.New("user already exists with this email")
	ErrRepositoryFailure    = errors.New("repository operation failed")
)
