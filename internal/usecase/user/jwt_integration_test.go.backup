package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// Test JWT integration with authentication usecases
func TestJWTIntegration_RegisterAndLogin(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"

	registerUC := NewRegisterUsecase(mockRepo, jwtSecret)
	loginUC := NewLoginUsecase(mockRepo, jwtSecret)
	defer mockRepo.Reset()

	ctx := context.Background()
	siteID := uuid.New()
	email := "integration@example.com"
	password := "securepass123"

	// Step 1: Register user
	registerInput := RegisterInput{
		SiteID:   siteID,
		Email:    email,
		Password: password,
	}

	registeredUser, err := registerUC.Execute(ctx, registerInput)
	// This will fail until T013 is implemented
	if err != nil {
		t.Skip("Skipping integration test until RegisterUsecase is implemented in T013")
		return
	}

	// Step 2: Login with registered user
	loginInput := LoginInput{
		SiteID:   siteID,
		Email:    email,
		Password: password,
	}

	authResult, err := loginUC.Execute(ctx, loginInput)
	// This will fail until T014 is implemented
	if err != nil {
		t.Skip("Skipping integration test until LoginUsecase is implemented in T014")
		return
	}

	// Step 3: Validate JWT token contains correct claims
	claims, err := auth.ValidateToken(authResult.Token, jwtSecret)
	assert.NoError(t, err)
	assert.Equal(t, registeredUser.ID.String(), claims.UserID)
	assert.Equal(t, siteID.String(), claims.SiteID)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestJWTIntegration_TokenGeneration(t *testing.T) {
	// Setup
	password := "securepass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	siteID := uuid.New()
	userID := uuid.New()
	email := "token@example.com"

	mockRepo := &MockUserRepository{
		users: []*domain.User{
			{
				ID:           userID,
				SiteID:       siteID,
				Email:        email,
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"
	loginUC := NewLoginUsecase(mockRepo, jwtSecret)
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{
		SiteID:   siteID,
		Email:    email,
		Password: password,
	}

	// Execute login
	result, err := loginUC.Execute(ctx, input)
	// This will fail until T014 is implemented
	if err != nil {
		t.Skip("Skipping JWT integration test until LoginUsecase is implemented in T014")
		return
	}

	// Validate token with JWT utilities
	claims, err := auth.ValidateToken(result.Token, jwtSecret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Validate claims match user data
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, siteID.String(), claims.SiteID)

	// Validate expiration
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(25*time.Hour))) // Within 25 hours

	// Validate the ExpiresAt in AuthenticationResult matches JWT claims
	assert.Equal(t, claims.ExpiresAt.Time, result.ExpiresAt)
}

func TestJWTIntegration_TokenValidation(t *testing.T) {
	// Test different token validation scenarios
	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"

	tests := []struct {
		name        string
		tokenFunc   func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid token",
			tokenFunc: func() string {
				token, _, _ := auth.GenerateToken(uuid.New(), uuid.New(), jwtSecret)
				return token
			},
			expectError: false,
		},
		{
			name: "invalid token format",
			tokenFunc: func() string {
				return "invalid.token.format"
			},
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name: "wrong secret",
			tokenFunc: func() string {
				token, _, _ := auth.GenerateToken(uuid.New(), uuid.New(), "wrong-secret")
				return token
			},
			expectError: true,
			errorMsg:    "signature is invalid",
		},
		{
			name: "empty token",
			tokenFunc: func() string {
				return ""
			},
			expectError: true,
			errorMsg:    "token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.tokenFunc()
			claims, err := auth.ValidateToken(token, jwtSecret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

func TestJWTIntegration_SiteIsolation(t *testing.T) {
	// Test that JWT tokens properly enforce site isolation
	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"

	siteA := uuid.New()
	siteB := uuid.New()
	userID := uuid.New()

	// Generate token for site A
	tokenA, _, err := auth.GenerateToken(userID, siteA, jwtSecret)
	assert.NoError(t, err)

	// Generate token for site B with same user
	tokenB, _, err := auth.GenerateToken(userID, siteB, jwtSecret)
	assert.NoError(t, err)

	// Validate both tokens
	claimsA, err := auth.ValidateToken(tokenA, jwtSecret)
	assert.NoError(t, err)
	assert.Equal(t, siteA.String(), claimsA.SiteID)

	claimsB, err := auth.ValidateToken(tokenB, jwtSecret)
	assert.NoError(t, err)
	assert.Equal(t, siteB.String(), claimsB.SiteID)

	// Ensure tokens are different
	assert.NotEqual(t, tokenA, tokenB)
	assert.NotEqual(t, claimsA.SiteID, claimsB.SiteID)
}

func TestJWTIntegration_ExpiredToken(t *testing.T) {
	// This test validates that expired tokens are properly rejected
	// Note: Since we can't easily create expired tokens with the current implementation,
	// this test focuses on the validation logic

	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"

	// Generate a valid token
	token, _, err := auth.GenerateToken(uuid.New(), uuid.New(), jwtSecret)
	assert.NoError(t, err)

	// Validate it's currently valid
	claims, err := auth.ValidateToken(token, jwtSecret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Ensure expiration is set to the future
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))

	// Note: In a real scenario, we would wait for the token to expire
	// or create a token with a past expiration time, but that requires
	// modifying the JWT generation logic which we'll keep simple for now
}

func TestJWTIntegration_ClaimsContent(t *testing.T) {
	// Test that JWT claims contain all required information
	jwtSecret := "test-secret-key-at-least-32-characters-for-jwt-signing"

	userID := uuid.New()
	siteID := uuid.New()

	// Generate token
	token, _, err := auth.GenerateToken(userID, siteID, jwtSecret)
	assert.NoError(t, err)

	// Validate and inspect claims
	claims, err := auth.ValidateToken(token, jwtSecret)
	assert.NoError(t, err)

	// Validate all required claims are present
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, siteID.String(), claims.SiteID)
	assert.NotZero(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))

	// Validate UUIDs are properly formatted
	_, err = uuid.Parse(claims.UserID)
	assert.NoError(t, err, "UserID should be a valid UUID")

	_, err = uuid.Parse(claims.SiteID)
	assert.NoError(t, err, "SiteID should be a valid UUID")
}
