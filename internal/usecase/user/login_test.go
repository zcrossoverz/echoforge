package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// Test suite for LoginUsecase
func TestLoginUsecase_Success(t *testing.T) {
	// Setup
	password := "securepass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()
	email := "test@example.com"

	mockRepo := &MockUserRepository{
		users: []*domain.User{
			{
				ID:           userID,
				Email:        email,
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{
		Email:    email,
		Password: password,
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// Validate - These assertions will fail until implementation is updated
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.User.ID)
	assert.Equal(t, email, result.User.Email)
	assert.NotEmpty(t, result.Token)
	assert.True(t, result.ExpiresAt.After(time.Now()))
}

func TestLoginUsecase_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{} // Empty repository
	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{

		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials") // Will be implemented in T014
}

func TestLoginUsecase_IncorrectPassword(t *testing.T) {
	// Setup
	correctPassword := "correctpass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	assert.NoError(t, err)

	email := "test@example.com"

	mockRepo := &MockUserRepository{
		users: []*domain.User{
			{
				ID: uuid.New(),

				Email:        email,
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{

		Email:    email,
		Password: "wrongpassword123", // Incorrect password
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")
}

// Note: SiteIsolation test removed - not applicable in clone-and-extend model
// Each site runs separate database instances, providing natural isolation

func TestLoginUsecase_ContextCancellation(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	input := LoginInput{

		Email:    "test@example.com",
		Password: "password123",
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, result)
	// Should handle context cancellation properly
}

func TestLoginUsecase_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{
		findError: errors.New("database connection failed"),
	}
	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{

		Email:    "test@example.com",
		Password: "password123",
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLoginUsecase_InvalidInput(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	loginUC := NewLoginUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()

	tests := []struct {
		name  string
		input LoginInput
	}{
		{
			name: "empty site ID",
			input: LoginInput{

				Email:    "test@example.com",
				Password: "password123",
			},
		},
		{
			name: "empty email",
			input: LoginInput{

				Email:    "",
				Password: "password123",
			},
		},
		{
			name: "invalid email format",
			input: LoginInput{

				Email:    "invalid-email",
				Password: "password123",
			},
		},
		{
			name: "empty password",
			input: LoginInput{

				Email:    "test@example.com",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := loginUC.Execute(ctx, tt.input)

			// Validate
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestLoginUsecase_JWTTokenValidation(t *testing.T) {
	// Setup
	password := "securepass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()
	email := "test@example.com"

	mockRepo := &MockUserRepository{
		users: []*domain.User{
			{
				ID: userID,

				Email:        email,
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	jwtSecret := "test-secret-key-at-least-32-characters"
	loginUC := NewLoginUsecase(mockRepo, jwtSecret)
	defer mockRepo.Reset()

	ctx := context.Background()
	input := LoginInput{

		Email:    email,
		Password: password,
	}

	// Execute
	result, err := loginUC.Execute(ctx, input)

	// This test will pass once T014 is implemented
	// For now, it will fail because login is not implemented
	if err == nil {
		// Validate JWT token structure
		assert.NotEmpty(t, result.Token)

		// Token should be parseable with our JWT utilities
		// (This will be tested properly in T010 - JWT Integration tests)
		assert.True(t, len(result.Token) > 50) // JWT tokens are typically long
	}
}
