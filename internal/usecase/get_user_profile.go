package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Profile request and response DTOs
type GetProfileRequest struct {
	Token string `json:"token" validate:"required"`
}

type GetProfileResponse struct {
	User *UserProfileDTO `json:"user"`
}

// Extended user profile DTO with additional fields
type UserProfileDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	// Future: Add profile fields like name, avatar, etc.
}

// GetUserProfileUseCase handles user profile retrieval business logic
type GetUserProfileUseCase struct {
	userRepo    domain.UserRepository
	authService domain.AuthService
}

// NewGetUserProfileUseCase creates a new get user profile use case
func NewGetUserProfileUseCase(
	userRepo domain.UserRepository,
	authService domain.AuthService,
) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		userRepo:    userRepo,
		authService: authService,
	}
}

// Execute retrieves user profile information using JWT token validation
func (uc *GetUserProfileUseCase) Execute(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	// Input validation
	if req == nil {
		return nil, errors.New("profile request is required")
	}

	if req.Token == "" {
		return nil, domain.ErrAuthorizationRequired
	}

	// Validate token and extract user ID
	userID, err := uc.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		// Map specific token errors for proper HTTP status codes
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

	// Handle case where user no longer exists
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Build profile DTO
	profileDTO := &UserProfileDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &GetProfileResponse{
		User: profileDTO,
	}, nil
}

// ExecuteWithToken retrieves profile directly with token string (convenience method)
func (uc *GetUserProfileUseCase) ExecuteWithToken(ctx context.Context, tokenString string) (*GetProfileResponse, error) {
	req := &GetProfileRequest{
		Token: tokenString,
	}

	return uc.Execute(ctx, req)
}

// GetProfileByUserID retrieves a user profile by user ID (internal use)
// This bypasses token validation and is intended for internal service calls
func (uc *GetUserProfileUseCase) GetProfileByUserID(ctx context.Context, userID string) (*UserProfileDTO, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Find user by ID
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Build profile DTO
	profileDTO := &UserProfileDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return profileDTO, nil
}

// ValidateProfileRequest validates the profile request structure
func (uc *GetUserProfileUseCase) ValidateProfileRequest(req *GetProfileRequest) error {
	if req == nil {
		return errors.New("profile request is required")
	}

	if req.Token == "" {
		return domain.ErrAuthorizationRequired
	}

	return nil
}

// ExtractUserIDFromToken validates a token and returns the user ID without database lookup
func (uc *GetUserProfileUseCase) ExtractUserIDFromToken(ctx context.Context, tokenString string) (string, error) {
	if tokenString == "" {
		return "", domain.ErrAuthorizationRequired
	}

	// Validate token and extract user ID
	userID, err := uc.authService.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// CheckTokenValidity validates a token without retrieving user data
func (uc *GetUserProfileUseCase) CheckTokenValidity(ctx context.Context, tokenString string) error {
	if tokenString == "" {
		return domain.ErrAuthorizationRequired
	}

	_, err := uc.authService.ValidateToken(ctx, tokenString)
	return err
}
