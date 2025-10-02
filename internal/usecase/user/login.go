package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// LoginInput represents the input for user login
type LoginInput struct {
	SiteID   uuid.UUID `json:"site_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email,max=320"`
	Password string    `json:"password" validate:"required,max=128"`
}

// AuthenticationResult represents the result of successful authentication
type AuthenticationResult struct {
	User      *domain.User `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// LoginUsecase defines the interface for user login business logic
type LoginUsecase interface {
	Execute(ctx context.Context, input LoginInput) (*AuthenticationResult, error)
}

// LoginUsecaseImpl implements LoginUsecase
type LoginUsecaseImpl struct {
	userRepo  domain.UserRepository
	jwtSecret string
}

// NewLoginUsecase creates a new LoginUsecase instance
func NewLoginUsecase(userRepo domain.UserRepository, jwtSecret string) LoginUsecase {
	return &LoginUsecaseImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Execute performs user authentication with JWT token generation
func (uc *LoginUsecaseImpl) Execute(ctx context.Context, input LoginInput) (*AuthenticationResult, error) {
	// Validate context
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate JWT secret
	if uc.jwtSecret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}

	// Validate input using go-playground/validator
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Find user by email and site (multi-tenant isolation)
	user, err := uc.userRepo.FindByEmail(ctx, input.SiteID, input.Email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token using auth utilities
	token, expiresAt, err := auth.GenerateToken(user.ID, user.SiteID, uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return authentication result
	result := &AuthenticationResult{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return result, nil
}
