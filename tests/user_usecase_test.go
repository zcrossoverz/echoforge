package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)
	ctx := context.Background()

	siteID := uuid.New()
	email := "new@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	createdUser, err := userUseCase.CreateUser(ctx, siteID, email, passwordHash)

	assert.NoError(t, err)
	require.NotNil(t, createdUser)
	assert.Equal(t, siteID, createdUser.SiteID)
	assert.Equal(t, email, createdUser.Email)
	assert.Equal(t, passwordHash, createdUser.PasswordHash)
	assert.NotEqual(t, uuid.Nil, createdUser.ID)
	assert.WithinDuration(t, time.Now(), createdUser.CreatedAt, 2*time.Second)
	assert.Equal(t, 1, mockRepo.CallCount("Create"))
}

func TestUserUseCase_CreateUser_InvalidEmail(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)
	ctx := context.Background()

	siteID := uuid.New()
	invalidEmail := "invalid-email"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	user, err := userUseCase.CreateUser(ctx, siteID, invalidEmail, passwordHash)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid email format")
	assert.Equal(t, 0, mockRepo.CallCount("Create")) // Should not reach repository
}

func TestUserUseCase_CreateUser_WeakPasswordHash(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)
	ctx := context.Background()

	siteID := uuid.New()
	email := "valid@example.com"
	weakPasswordHash := "short_hash"

	user, err := userUseCase.CreateUser(ctx, siteID, email, weakPasswordHash)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password hash too short")
	assert.Equal(t, 0, mockRepo.CallCount("Create")) // Should not reach repository
}

func TestUserUseCase_CreateUser_DuplicateEmail(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)
	ctx := context.Background()

	siteID := uuid.New()
	email := "duplicate@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Create first user
	user1, err := userUseCase.CreateUser(ctx, siteID, email, passwordHash)
	require.NoError(t, err)
	require.NotNil(t, user1)

	// Try to create second user with same email in same site
	user2, err := userUseCase.CreateUser(ctx, siteID, email, passwordHash)

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	assert.Nil(t, user2)
	assert.Equal(t, 1, mockRepo.CallCount("Create"))      // Only first creation reaches repository
	assert.Equal(t, 2, mockRepo.CallCount("FindByEmail")) // Both check for existing user
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)
	ctx := context.Background()

	siteID := uuid.New()
	email := "existing@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Create user first
	createdUser, err := userUseCase.CreateUser(ctx, siteID, email, passwordHash)
	require.NoError(t, err)

	// Retrieve user by email
	foundUser, err := userUseCase.GetUserByEmail(ctx, siteID, email)

	assert.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.SiteID, foundUser.SiteID)
	assert.Equal(t, createdUser.Email, foundUser.Email)
	assert.Equal(t, createdUser.PasswordHash, foundUser.PasswordHash)
	assert.Equal(t, 2, mockRepo.CallCount("FindByEmail")) // CreateUser checks availability + GetUserByEmail retrieves
}
