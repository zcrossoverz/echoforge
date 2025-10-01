package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user within a site context
type User struct {
	ID           uuid.UUID `json:"id"`
	SiteID       uuid.UUID `json:"site_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validation constants
const (
	MaxEmailLength    = 255
	MinPasswordLength = 60 // bcrypt hash length
)

// Domain errors
var (
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrEmailTooLong         = errors.New("email exceeds maximum length")
	ErrPasswordHashTooShort = errors.New("password hash too short")
	ErrRequiredField        = errors.New("required field is empty")
	ErrUserAlreadyExists    = errors.New("user already exists with this email in site")
	ErrRepositoryFailure    = errors.New("repository operation failed")
)

// NewUser creates a new User entity with validation
func NewUser(siteID uuid.UUID, email, passwordHash string) (*User, error) {
	now := time.Now()

	user := &User{
		ID:           uuid.New(),
		SiteID:       siteID,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// Validate performs business rule validation
func (u *User) Validate() error {
	// Required fields
	if u.ID == uuid.Nil {
		return fmt.Errorf("ID: %w", ErrRequiredField)
	}
	if u.SiteID == uuid.Nil {
		return fmt.Errorf("SiteID: %w", ErrRequiredField)
	}
	if u.Email == "" {
		return fmt.Errorf("Email: %w", ErrRequiredField)
	}
	if u.PasswordHash == "" {
		return fmt.Errorf("PasswordHash: %w", ErrRequiredField)
	}

	// Email validation
	if len(u.Email) > MaxEmailLength {
		return fmt.Errorf("Email: %w (%d > %d)", ErrEmailTooLong, len(u.Email), MaxEmailLength)
	}
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("Email: %w", ErrInvalidEmail)
	}

	// Password hash validation
	if len(u.PasswordHash) < MinPasswordLength {
		return fmt.Errorf("PasswordHash: %w (%d < %d)", ErrPasswordHashTooShort, len(u.PasswordHash), MinPasswordLength)
	}

	return nil
}

// IsValid returns true if entity passes all validation rules
func (u *User) IsValid() bool {
	return u.Validate() == nil
}

// UserRepository defines the persistence contract for User entities
type UserRepository interface {
	// Create persists a new user entity
	Create(ctx context.Context, user *User) error

	// FindByEmail retrieves a user by email within specific site
	FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error)
}
