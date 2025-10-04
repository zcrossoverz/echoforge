package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Registration request and response DTOs
type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type RegisterUserResponse struct {
	User  *UserDTO  `json:"user"`
	Token *TokenDTO `json:"token"`
}

type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TokenDTO struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// UserRegistrationUseCase handles user registration business logic
type UserRegistrationUseCase struct {
	userRepo        domain.UserRepository
	authService     domain.AuthService
	passwordService domain.PasswordService
}

// NewUserRegistrationUseCase creates a new user registration use case
func NewUserRegistrationUseCase(
	userRepo domain.UserRepository,
	authService domain.AuthService,
	passwordService domain.PasswordService,
) *UserRegistrationUseCase {
	return &UserRegistrationUseCase{
		userRepo:        userRepo,
		authService:     authService,
		passwordService: passwordService,
	}
}

// Execute performs user registration with business rule validation
func (uc *UserRegistrationUseCase) Execute(ctx context.Context, req *RegisterUserRequest) (*RegisterUserResponse, error) {
	// Input validation
	if req == nil {
		return nil, errors.New("registration request is required")
	}

	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Check if user already exists (email uniqueness constraint)
	existingUser, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Validate password strength
	if err := uc.passwordService.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user, err := domain.NewUser(req.Email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Persist user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate authentication token
	authToken, err := uc.authService.GenerateToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	// Build response DTOs
	userDTO := &UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	tokenDTO := &TokenDTO{
		Token:     authToken.Token,
		ExpiresAt: authToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &RegisterUserResponse{
		User:  userDTO,
		Token: tokenDTO,
	}, nil
}

// ValidateRegistrationRequest validates the registration request structure
func (uc *UserRegistrationUseCase) ValidateRegistrationRequest(req *RegisterUserRequest) error {
	if req == nil {
		return errors.New("registration request is required")
	}

	// Email validation (domain rules)
	if req.Email == "" {
		return errors.New("email is required")
	}

	// Create a temporary user to leverage domain validation
	tempUser := &domain.User{
		Email: req.Email,
	}

	// Use domain validation for email format and length
	if len(req.Email) > domain.MaxEmailLength {
		return domain.ErrEmailTooLong
	}

	// Validate email format using domain logic
	if err := tempUser.Validate(); err != nil {
		// Extract just the email validation error
		if errors.Is(err, domain.ErrInvalidEmail) {
			return domain.ErrInvalidEmail
		}
	}

	// Password validation (delegated to password service)
	if err := uc.passwordService.ValidatePassword(req.Password); err != nil {
		return err
	}

	return nil
}

// CheckEmailAvailability checks if an email is available for registration
func (uc *UserRegistrationUseCase) CheckEmailAvailability(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, errors.New("email is required")
	}

	exists, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email availability: %w", err)
	}

	return !exists, nil
}
