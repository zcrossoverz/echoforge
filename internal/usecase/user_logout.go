package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Logout request and response DTOs
type LogoutRequest struct {
	Token string `json:"token" validate:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// UserLogoutUseCase handles user logout business logic
type UserLogoutUseCase struct {
	authService domain.AuthService
}

// NewUserLogoutUseCase creates a new user logout use case
func NewUserLogoutUseCase(authService domain.AuthService) *UserLogoutUseCase {
	return &UserLogoutUseCase{
		authService: authService,
	}
}

// Execute performs user logout by blacklisting the token
func (uc *UserLogoutUseCase) Execute(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	// Input validation
	if req == nil {
		return nil, errors.New("logout request is required")
	}

	if req.Token == "" {
		return nil, domain.ErrAuthorizationRequired
	}

	// First validate the token to ensure it's a valid JWT
	// This prevents adding invalid tokens to the blacklist
	_, err := uc.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		// If token is already invalid, treat as successful logout
		// This prevents errors on duplicate logout attempts
		if errors.Is(err, domain.ErrTokenExpired) {
			return &LogoutResponse{
				Message: "Successfully logged out (token was already expired)",
				Success: true,
			}, nil
		}
		if errors.Is(err, domain.ErrTokenBlacklisted) {
			return &LogoutResponse{
				Message: "Successfully logged out (token was already blacklisted)",
				Success: true,
			}, nil
		}
		if errors.Is(err, domain.ErrTokenInvalid) {
			return &LogoutResponse{
				Message: "Successfully logged out (token was invalid)",
				Success: true,
			}, nil
		}

		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Add token to blacklist
	if err := uc.authService.BlacklistToken(ctx, req.Token); err != nil {
		return nil, fmt.Errorf("failed to blacklist token: %w", err)
	}

	return &LogoutResponse{
		Message: "Successfully logged out",
		Success: true,
	}, nil
}

// ValidateLogoutRequest validates the logout request structure
func (uc *UserLogoutUseCase) ValidateLogoutRequest(req *LogoutRequest) error {
	if req == nil {
		return errors.New("logout request is required")
	}

	if req.Token == "" {
		return domain.ErrAuthorizationRequired
	}

	return nil
}

// ExecuteWithToken performs logout directly with a token string (convenience method)
func (uc *UserLogoutUseCase) ExecuteWithToken(ctx context.Context, tokenString string) (*LogoutResponse, error) {
	req := &LogoutRequest{
		Token: tokenString,
	}

	return uc.Execute(ctx, req)
}

// LogoutAll logs out a user from all devices by blacklisting all their tokens
// Note: This requires additional infrastructure to track user tokens
// For now, this is a placeholder for future implementation
func (uc *UserLogoutUseCase) LogoutAll(ctx context.Context, userID string) error {
	// TODO: Implement when we have user token tracking
	// This would require:
	// 1. A way to track all active tokens for a user
	// 2. Blacklisting all tokens for the user
	// 3. Updating the user's "logged out at" timestamp

	return errors.New("logout all functionality not yet implemented")
}

// IsTokenBlacklisted checks if a token has been blacklisted (logout verification)
func (uc *UserLogoutUseCase) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	if tokenString == "" {
		return false, errors.New("token is required")
	}

	return uc.authService.IsTokenBlacklisted(ctx, tokenString)
}

// CleanupExpiredTokens removes expired tokens from the blacklist
// This should be called periodically to prevent the blacklist from growing indefinitely
func (uc *UserLogoutUseCase) CleanupExpiredTokens(ctx context.Context) error {
	return uc.authService.CleanupExpiredTokens(ctx)
}
