package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Authentication request and response DTOs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User  *UserDTO  `json:"user"`
	Token *TokenDTO `json:"token"`
}

// UserAuthenticationUseCase handles user login business logic
type UserAuthenticationUseCase struct {
	userRepo        domain.UserRepository
	authService     domain.AuthService
	passwordService domain.PasswordService
}

// NewUserAuthenticationUseCase creates a new user authentication use case
func NewUserAuthenticationUseCase(
	userRepo domain.UserRepository,
	authService domain.AuthService,
	passwordService domain.PasswordService,
) *UserAuthenticationUseCase {
	return &UserAuthenticationUseCase{
		userRepo:        userRepo,
		authService:     authService,
		passwordService: passwordService,
	}
}

// Execute performs user authentication with credential validation
func (uc *UserAuthenticationUseCase) Execute(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Input validation
	if req == nil {
		return nil, errors.New("login request is required")
	}

	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// User not found - return generic error to prevent email enumeration
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if !uc.passwordService.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
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

	return &LoginResponse{
		User:  userDTO,
		Token: tokenDTO,
	}, nil
}

// ValidateLoginRequest validates the login request structure
func (uc *UserAuthenticationUseCase) ValidateLoginRequest(req *LoginRequest) error {
	if req == nil {
		return errors.New("login request is required")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	// Basic email format validation (more lenient than registration)
	if len(req.Email) > domain.MaxEmailLength {
		return domain.ErrEmailTooLong
	}

	// Validate email format using domain logic
	tempUser := &domain.User{
		Email: req.Email,
	}

	if err := tempUser.Validate(); err != nil {
		if errors.Is(err, domain.ErrInvalidEmail) {
			return domain.ErrInvalidEmail
		}
	}

	return nil
}

// AuthenticateWithToken validates an authentication token and returns user info
func (uc *UserAuthenticationUseCase) AuthenticateWithToken(ctx context.Context, tokenString string) (*UserDTO, error) {
	if tokenString == "" {
		return nil, domain.ErrAuthorizationRequired
	}

	// Validate token and extract user ID
	userID, err := uc.authService.ValidateToken(ctx, tokenString)
	if err != nil {
		// Map specific token errors
		if errors.Is(err, domain.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		if errors.Is(err, domain.ErrTokenBlacklisted) {
			return nil, domain.ErrTokenBlacklisted
		}
		return nil, domain.ErrTokenInvalid
	}

	// Find user by ID
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Build user DTO
	userDTO := &UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return userDTO, nil
}

// RefreshToken generates a new token from a valid existing token
func (uc *UserAuthenticationUseCase) RefreshToken(ctx context.Context, currentToken string) (*TokenDTO, error) {
	if currentToken == "" {
		return nil, domain.ErrAuthorizationRequired
	}

	// Refresh token using auth service
	authToken, err := uc.authService.RefreshToken(ctx, currentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Build token DTO
	tokenDTO := &TokenDTO{
		Token:     authToken.Token,
		ExpiresAt: authToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return tokenDTO, nil
}
