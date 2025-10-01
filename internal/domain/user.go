// internal/domain/user.go
package domain

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User entity (value object - immutable after create)
type User struct {
	ID           uuid.UUID `json:"id" validate:"required,uuid"` // Add 'uuid' tag cho stricter check
	Email        string    `json:"email" validate:"required,email"`
	PasswordHash string    `json:"-"` // Không expose raw hash
	Role         string    `json:"role" validate:"omitempty,oneof=user admin"`
	CreatedAt    time.Time `json:"created_at" validate:"omitempty"` // Optional tag
}

// NewUser: Constructor với hashing trước, validate full entity sau (fix nil ID)
func NewUser(email, password string) (*User, error) {
	// Hash password first (business invariant)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Construct full entity
	user := &User{
		ID:           uuid.New(), // Set ID trước validate
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user", // Default
		CreatedAt:    time.Now(),
	}

	// Validate full entity (ID now set, no nil)
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(user); err != nil {
		return nil, err.(validator.ValidationErrors) // Or wrap to domain.ErrInvalidEmail
	}

	return user, nil
}

// CheckPassword: Giữ nguyên
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UserRepository: Giữ nguyên
type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
}

// Domain errors: Giữ nguyên, add nếu cần
var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidEmail  = errors.New("invalid email")
	ErrWrongPassword = errors.New("wrong password")
)
