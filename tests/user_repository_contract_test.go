package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Note: MockUserRepository is implemented in mock_user_repository.go

// Contract Tests
func TestUserRepository_Create_Success(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	user, err := domain.NewUser("user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repo.Create(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, 1, repo.CallCount("Create"))

	// Verify user can be retrieved
	retrieved, err := repo.FindByEmail(ctx, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create first user
	user1, err := domain.NewUser("user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same email
	user2, err := domain.NewUser("user@example.com", "different_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, user2)

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestUserRepository_Create_SameEmailDifferentSites(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	siteA := uuid.New()
	siteB := uuid.New()
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Create user in site A
	userA, err := domain.NewUser(email, passwordHash)
	require.NoError(t, err)
	err = repo.Create(ctx, userA)
	require.NoError(t, err)

	// Create user with same email in site B (should succeed)
	userB, err := domain.NewUser(siteB, email, passwordHash)
	require.NoError(t, err)
	err = repo.Create(ctx, userB)

	assert.NoError(t, err) // No conflict between different sites
}

func TestUserRepository_Create_ContextCancellation(t *testing.T) {
	repo := NewMockUserRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	siteID := uuid.New()
	user, err := domain.NewUser(siteID, "user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repo.Create(ctx, user)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestUserRepository_Create_ContextTimeout(t *testing.T) {
	repo := NewMockUserRepository()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Force timeout

	siteID := uuid.New()
	user, err := domain.NewUser(siteID, "user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repo.Create(ctx, user)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()
	siteID := uuid.New()

	// Create and store user
	user, err := domain.NewUser(siteID, "user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Find user by email
	retrieved, err := repo.FindByEmail(ctx, siteID, "user@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.SiteID, retrieved.SiteID)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, 1, repo.CallCount("FindByEmail"))
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()
	siteID := uuid.New()

	// Try to find non-existent user
	user, err := repo.FindByEmail(ctx, siteID, "nonexistent@example.com")

	assert.NoError(t, err) // Not found is not an error
	assert.Nil(t, user)    // Returns nil for not found
}

func TestUserRepository_FindByEmail_SiteIsolation(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	siteA := uuid.New()
	siteB := uuid.New()
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Create user in site A
	userA, err := domain.NewUser(siteA, email, passwordHash)
	require.NoError(t, err)
	err = repo.Create(ctx, userA)
	require.NoError(t, err)

	// Try to find user from site B (should not find)
	user, err := repo.FindByEmail(ctx, siteB, email)
	assert.NoError(t, err)
	assert.Nil(t, user) // Site isolation prevents cross-site access

	// Find user from correct site A (should find)
	user, err = repo.FindByEmail(ctx, siteA, email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userA.ID, user.ID)
}

func TestUserRepository_FindByEmail_ContextCancellation(t *testing.T) {
	repo := NewMockUserRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	siteID := uuid.New()
	user, err := repo.FindByEmail(ctx, siteID, "user@example.com")

	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, user)
}

func TestUserRepository_Contract_Interface(t *testing.T) {
	// Verify MockUserRepository implements domain.UserRepository interface
	var _ domain.UserRepository = (*MockUserRepository)(nil)

	// This test will fail to compile if interface is not properly implemented
	repo := NewMockUserRepository()
	assert.NotNil(t, repo)
}
