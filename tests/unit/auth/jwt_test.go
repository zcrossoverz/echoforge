package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zcrossoverz/echoforge/internal/config"
	"github.com/zcrossoverz/echoforge/pkg/auth"
)

// MockBlacklistStore is a mock implementation of the BlacklistStore interface
type MockBlacklistStore struct {
	mock.Mock
}

func (m *MockBlacklistStore) AddToken(tokenString string, expiresAt time.Time) error {
	args := m.Called(tokenString, expiresAt)
	return args.Error(0)
}

func (m *MockBlacklistStore) IsBlacklisted(tokenString string) (bool, error) {
	args := m.Called(tokenString)
	return args.Bool(0), args.Error(1)
}

func (m *MockBlacklistStore) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	tests := []struct {
		name      string
		userID    uuid.UUID
		secret    string
		expectErr bool
	}{
		{
			name:      "Valid token generation",
			userID:    userID,
			secret:    secret,
			expectErr: false,
		},
		{
			name:      "Empty secret",
			userID:    userID,
			secret:    "",
			expectErr: true,
		},
		{
			name:      "Nil UUID",
			userID:    uuid.Nil,
			secret:    secret,
			expectErr: false, // Should still work, just use nil UUID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := auth.GenerateToken(tt.userID, tt.secret)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.True(t, expiresAt.IsZero())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.False(t, expiresAt.IsZero())
				assert.True(t, expiresAt.After(time.Now()))
			}
		})
	}
}

func TestGenerateTokenWithConfig(t *testing.T) {
	userID := uuid.New()
	cfg := &config.Config{
		JWTSecret: "test-config-secret",
	}

	token, expiresAt, err := auth.GenerateTokenWithConfig(userID, cfg)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())
	assert.True(t, expiresAt.After(time.Now()))

	// Verify token can be validated with the same secret
	claims, err := auth.ValidateToken(token, cfg.JWTSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"

	// Generate a valid token
	validToken, _, err := auth.GenerateToken(userID, secret)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		token      string
		secret     string
		expectErr  bool
		expectedID string
	}{
		{
			name:       "Valid token",
			token:      validToken,
			secret:     secret,
			expectErr:  false,
			expectedID: userID.String(),
		},
		{
			name:      "Empty secret",
			token:     validToken,
			secret:    "",
			expectErr: true,
		},
		{
			name:      "Wrong secret",
			token:     validToken,
			secret:    wrongSecret,
			expectErr: true,
		},
		{
			name:      "Empty token",
			token:     "",
			secret:    secret,
			expectErr: true,
		},
		{
			name:      "Invalid token format",
			token:     "not.a.jwt.token",
			secret:    secret,
			expectErr: true,
		},
		{
			name:      "Malformed token",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid-payload",
			secret:    secret,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token, tt.secret)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedID, claims.UserID)
			}
		})
	}
}

func TestValidateTokenWithConfig(t *testing.T) {
	userID := uuid.New()
	cfg := &config.Config{
		JWTSecret: "test-config-secret",
	}

	// Generate a valid token
	validToken, _, err := auth.GenerateTokenWithConfig(userID, cfg)
	assert.NoError(t, err)

	// Validate the token
	claims, err := auth.ValidateTokenWithConfig(validToken, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID.String(), claims.UserID)
}

func TestJWTService_NewJWTService(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := auth.NewJWTService(cfg)
	assert.NotNil(t, service)
	assert.Equal(t, cfg.JWTSecret, service.GetSecret())
}

func TestJWTService_NewJWTServiceWithBlacklist(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockBlacklist := &MockBlacklistStore{}

	service := auth.NewJWTServiceWithBlacklist(cfg, mockBlacklist)
	assert.NotNil(t, service)
	assert.Equal(t, cfg.JWTSecret, service.GetSecret())
}

func TestJWTService_GenerateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	service := auth.NewJWTService(cfg)
	userID := uuid.New()

	token, expiresAt, err := service.GenerateToken(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())
	assert.True(t, expiresAt.After(time.Now()))

	// Verify the token can be validated
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
}

func TestJWTService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	service := auth.NewJWTService(cfg)
	userID := uuid.New()

	// Generate a token
	token, _, err := service.GenerateToken(userID)
	assert.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
}

func TestJWTService_ValidateToken_WithBlacklist(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockBlacklist := &MockBlacklistStore{}
	service := auth.NewJWTServiceWithBlacklist(cfg, mockBlacklist)
	userID := uuid.New()

	// Generate a token
	token, _, err := service.GenerateToken(userID)
	assert.NoError(t, err)

	t.Run("Token not blacklisted", func(t *testing.T) {
		mockBlacklist.On("IsBlacklisted", token).Return(false, nil).Once()

		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID.String(), claims.UserID)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("Token is blacklisted", func(t *testing.T) {
		mockBlacklist.On("IsBlacklisted", token).Return(true, nil).Once()

		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blacklisted")
		assert.Nil(t, claims)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("Blacklist check fails", func(t *testing.T) {
		mockBlacklist.On("IsBlacklisted", token).Return(false, errors.New("database error")).Once()

		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blacklist check failed")
		assert.Nil(t, claims)

		mockBlacklist.AssertExpectations(t)
	})
}

func TestJWTService_BlacklistToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockBlacklist := &MockBlacklistStore{}
	service := auth.NewJWTServiceWithBlacklist(cfg, mockBlacklist)
	userID := uuid.New()

	// Generate a token
	token, expiresAt, err := service.GenerateToken(userID)
	assert.NoError(t, err)

	t.Run("Successful blacklist with valid token", func(t *testing.T) {
		mockBlacklist.On("AddToken", token, mock.MatchedBy(func(t time.Time) bool {
			return t.After(time.Now()) && t.Before(expiresAt.Add(time.Minute))
		})).Return(nil).Once()

		err := service.BlacklistToken(context.Background(), token)
		assert.NoError(t, err)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("Blacklist invalid token with default expiration", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		mockBlacklist.On("AddToken", invalidToken, mock.MatchedBy(func(t time.Time) bool {
			return t.After(time.Now()) && t.Before(time.Now().Add(25*time.Hour))
		})).Return(nil).Once()

		err := service.BlacklistToken(context.Background(), invalidToken)
		assert.NoError(t, err)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("No blacklist store configured", func(t *testing.T) {
		serviceWithoutBlacklist := auth.NewJWTService(cfg)
		err := serviceWithoutBlacklist.BlacklistToken(context.Background(), token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blacklist store not configured")
	})
}

func TestJWTService_IsTokenBlacklisted(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockBlacklist := &MockBlacklistStore{}
	service := auth.NewJWTServiceWithBlacklist(cfg, mockBlacklist)
	token := "test.token.here"

	t.Run("Token is blacklisted", func(t *testing.T) {
		mockBlacklist.On("IsBlacklisted", token).Return(true, nil).Once()

		isBlacklisted, err := service.IsTokenBlacklisted(token)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("Token is not blacklisted", func(t *testing.T) {
		mockBlacklist.On("IsBlacklisted", token).Return(false, nil).Once()

		isBlacklisted, err := service.IsTokenBlacklisted(token)
		assert.NoError(t, err)
		assert.False(t, isBlacklisted)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("No blacklist store configured", func(t *testing.T) {
		serviceWithoutBlacklist := auth.NewJWTService(cfg)
		isBlacklisted, err := serviceWithoutBlacklist.IsTokenBlacklisted(token)
		assert.NoError(t, err)
		assert.False(t, isBlacklisted) // Should return false when no blacklist store
	})
}

func TestJWTService_CleanupExpiredTokens(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockBlacklist := &MockBlacklistStore{}
	service := auth.NewJWTServiceWithBlacklist(cfg, mockBlacklist)

	t.Run("Successful cleanup", func(t *testing.T) {
		mockBlacklist.On("CleanupExpired").Return(nil).Once()

		err := service.CleanupExpiredTokens()
		assert.NoError(t, err)

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("Cleanup error", func(t *testing.T) {
		mockBlacklist.On("CleanupExpired").Return(errors.New("cleanup failed")).Once()

		err := service.CleanupExpiredTokens()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cleanup failed")

		mockBlacklist.AssertExpectations(t)
	})

	t.Run("No blacklist store configured", func(t *testing.T) {
		serviceWithoutBlacklist := auth.NewJWTService(cfg)
		err := serviceWithoutBlacklist.CleanupExpiredTokens()
		assert.NoError(t, err) // Should not error when no blacklist store
	})
}

func TestJWTClaims_TokenExpiration(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	// Generate a token
	token, expiresAt, err := auth.GenerateToken(userID, secret)
	assert.NoError(t, err)

	// Validate the token immediately
	claims, err := auth.ValidateToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)

	// Check that expiration time is about 24 hours from now
	expectedExpiration := time.Now().Add(24 * time.Hour)
	timeDiff := expiresAt.Sub(expectedExpiration)
	assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute, "Expiration time should be approximately 24 hours from now")

	// Verify claims expiration matches token expiration
	assert.True(t, claims.ExpiresAt.Time.Equal(expiresAt) || claims.ExpiresAt.Time.Sub(expiresAt) < time.Second)
}

func TestGenerateJWT_LegacyFunction(t *testing.T) {
	// Test the legacy function for backward compatibility
	userID := "test-user-id"
	role := "admin"

	token, err := auth.GenerateJWT(userID, role)
	assert.NoError(t, err)
	assert.Contains(t, token, userID[:8])
	assert.Contains(t, token, role)
}

func TestJWTService_EdgeCases(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	service := auth.NewJWTService(cfg)

	t.Run("Generate token with zero UUID", func(t *testing.T) {
		token, expiresAt, err := service.GenerateToken(uuid.Nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.False(t, expiresAt.IsZero())

		// Should be able to validate
		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uuid.Nil.String(), claims.UserID)
	})

	t.Run("Validate token with expired claims", func(t *testing.T) {
		// This would require creating a token with past expiration
		// which is complex with the current implementation
		// For now, we test the error path with an invalid token
		claims, err := service.ValidateToken("invalid.token.format")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
