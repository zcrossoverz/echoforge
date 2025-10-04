package usecase

import (
	"context"
	"fmt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// UserUseCase handles user-related business logic
type UserUseCase struct {
	userRepo domain.UserRepository
}

// NewUserUseCase creates a new UserUseCase instance
func NewUserUseCase(userRepo domain.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with validation and persistence
func (uc *UserUseCase) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	// Check context early
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate input parameters
	if email == "" {
		return nil, fmt.Errorf("invalid email: cannot be empty")
	}

	if passwordHash == "" {
		return nil, fmt.Errorf("invalid password hash: cannot be empty")
	}

	// Create domain entity (this will validate business rules)
	user, err := domain.NewUser(email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Check if user already exists (business rule enforcement)
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Persist the user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	// Check context early
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate input parameters
	if email == "" {
		return nil, fmt.Errorf("invalid email: cannot be empty")
	}

	// Retrieve user from repository
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// IsEmailAvailable checks if an email is available
func (uc *UserUseCase) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	// Check context
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	// Validate parameters
	if email == "" {
		return false, fmt.Errorf("invalid email: cannot be empty")
	}

	// Check if user exists
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email availability: %w", err)
	}

	return user == nil, nil
}
