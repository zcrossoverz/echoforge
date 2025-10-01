package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput represents the input for user registration
type RegisterInput struct {
	SiteID   uuid.UUID `json:"site_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email,max=320"`
	Password string    `json:"password" validate:"required,min=8,max=128"`
}

// RegisterUsecase defines the interface for user registration business logic
type RegisterUsecase interface {
	Execute(ctx context.Context, input RegisterInput) (*domain.User, error)
}

// RegisterUsecaseImpl implements RegisterUsecase
type RegisterUsecaseImpl struct {
	userRepo  domain.UserRepository
	jwtSecret string
}

// NewRegisterUsecase creates a new RegisterUsecase instance
func NewRegisterUsecase(userRepo domain.UserRepository, jwtSecret string) RegisterUsecase {
	return &RegisterUsecaseImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Execute performs user registration with validation and multi-tenant isolation
func (uc *RegisterUsecaseImpl) Execute(ctx context.Context, input RegisterInput) (*domain.User, error) {
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
	
	// Check if user already exists (multi-tenant: same email allowed across different sites)
	existingUser, err := uc.userRepo.FindByEmail(ctx, input.SiteID, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already exists for this site")
	}
	
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create new user
	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		SiteID:       input.SiteID,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	// Save user to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}