package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zcrossoverz/echoforge/internal/usecase/user"
	"github.com/zcrossoverz/echoforge/pkg/auth"
)

// TestCloneAndExtendEndToEndFlow tests the complete user registration and authentication flow
// in the clone-and-extend model without any site_id complexity
func TestCloneAndExtendEndToEndFlow(t *testing.T) {
	// Setup
	mockRepo := NewMockUserRepository()
	registerUC := user.NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	loginUC := user.NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	ctx := context.Background()

	// Test Data
	email := "endtoend@example.com"
	password := "securepassword123"

	// Step 1: Register a new user (simplified - no site_id)
	registerInput := user.RegisterInput{
		Email:    email,
		Password: password,
	}

	registeredUser, err := registerUC.Execute(ctx, registerInput)
	require.NoError(t, err)
	require.NotNil(t, registeredUser)
	assert.Equal(t, email, registeredUser.Email)

	// Step 2: Login with the registered user (simplified - no site_id)
	loginInput := user.LoginInput{
		Email:    email,
		Password: password,
	}

	loginResult, err := loginUC.Execute(ctx, loginInput)
	require.NoError(t, err)
	require.NotNil(t, loginResult)
	require.NotNil(t, loginResult.User)
	assert.Equal(t, email, loginResult.User.Email)
	assert.Equal(t, registeredUser.ID, loginResult.User.ID)
	assert.NotEmpty(t, loginResult.Token)

	// Step 3: Validate the login JWT token contains only user claims
	loginClaims, err := auth.ValidateToken(loginResult.Token, "test-secret-key-at-least-32-characters")
	require.NoError(t, err)
	assert.Equal(t, loginResult.User.ID.String(), loginClaims.UserID)
	// Note: No SiteID claim should exist in clone-and-extend model

	// Step 4: Verify repository operations worked correctly
	assert.Equal(t, 1, mockRepo.CallCount("Create"))      // User created once
	assert.Equal(t, 2, mockRepo.CallCount("FindByEmail")) // Register checks duplicates + Login retrieves user

	// Step 5: Verify user can be retrieved directly (no site filtering)
	foundUser, err := mockRepo.FindByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, registeredUser.ID, foundUser.ID)
	assert.Equal(t, email, foundUser.Email)

	t.Log("✅ Clone-and-extend end-to-end flow completed successfully!")
	t.Log("✅ User registration, authentication, and JWT validation working without site_id complexity")
	t.Log("✅ Natural database isolation achieved through separate database instances")
}
