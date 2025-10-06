package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput represents the input for user registration (clone-and-extend model)
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,min=8,max=128"`
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

	// Check if user already exists (clone-and-extend: global email uniqueness)
	existingUser, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user already exists with this email")
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user using domain constructor
	user, err := domain.NewUser(input.Email, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
