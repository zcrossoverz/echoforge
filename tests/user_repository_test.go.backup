package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zcrossoverz/echoforge/adapters/persistence"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// Test database configuration
const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "postgres"
	testDBPassword = "password"
	testDBName     = "echoforge_test"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Skip if no database available
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create test database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&persistence.GormUser{})
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		// Clean up test data
		db.Exec("DELETE FROM users")

		// Close connection
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

func TestUserRepository_Create_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	user, err := domain.NewUser(siteID, "user@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repo.Create(ctx, user)

	assert.NoError(t, err)

	// Verify user was persisted
	var count int64
	db.Model(&persistence.GormUser{}).Where("site_id = ? AND email = ?", siteID, "user@example.com").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestUserRepository_Create_DuplicateEmailSameSite_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	// Create first user
	user1, err := domain.NewUser(siteID, "duplicate@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same email in same site
	user2, err := domain.NewUser(siteID, "duplicate@example.com", "different_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, user2)

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestUserRepository_Create_SameEmailDifferentSites_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()

	siteA := uuid.New()
	siteB := uuid.New()
	email := "shared@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Create user in site A
	userA, err := domain.NewUser(siteA, email, passwordHash)
	require.NoError(t, err)
	err = repo.Create(ctx, userA)
	require.NoError(t, err)

	// Create user with same email in site B (should succeed)
	userB, err := domain.NewUser(siteB, email, passwordHash)
	require.NoError(t, err)
	err = repo.Create(ctx, userB)

	assert.NoError(t, err) // No conflict between different sites

	// Verify both users exist
	var count int64
	db.Model(&persistence.GormUser{}).Where("email = ?", email).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestUserRepository_FindByEmail_Success_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	// Create and store user
	originalUser, err := domain.NewUser(siteID, "find@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)
	err = repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Find user by email
	foundUser, err := repo.FindByEmail(ctx, siteID, "find@example.com")

	assert.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, originalUser.ID, foundUser.ID)
	assert.Equal(t, originalUser.SiteID, foundUser.SiteID)
	assert.Equal(t, originalUser.Email, foundUser.Email)
	assert.Equal(t, originalUser.PasswordHash, foundUser.PasswordHash)
}

func TestUserRepository_FindByEmail_NotFound_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	// Try to find non-existent user
	user, err := repo.FindByEmail(ctx, siteID, "nonexistent@example.com")

	assert.NoError(t, err) // Not found is not an error
	assert.Nil(t, user)    // Returns nil for not found
}

func TestUserRepository_FindByEmail_SiteIsolation_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()

	siteA := uuid.New()
	siteB := uuid.New()
	email := "isolated@example.com"
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

func TestUserRepository_ContextTimeout_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Force timeout

	siteID := uuid.New()
	user, err := domain.NewUser(siteID, "timeout@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repo.Create(ctx, user)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUserRepository_DatabaseTransactionRollback_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	// Start transaction
	tx := db.Begin()
	repoTx := persistence.NewUserRepository(tx)

	user, err := domain.NewUser(siteID, "rollback@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	require.NoError(t, err)

	err = repoTx.Create(ctx, user)
	require.NoError(t, err)

	// Rollback transaction
	tx.Rollback()

	// Verify user was not persisted
	foundUser, err := repo.FindByEmail(ctx, siteID, "rollback@example.com")
	assert.NoError(t, err)
	assert.Nil(t, foundUser) // Should not exist after rollback
}

// Benchmark tests for performance validation
func BenchmarkUserRepository_Create(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("bench%d@example.com", i)
		user, _ := domain.NewUser(siteID, email, "encrypted_password_hash_that_is_at_least_sixty_characters_long")
		repo.Create(ctx, user)
	}
}

func BenchmarkUserRepository_FindByEmail(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	repo := persistence.NewUserRepository(db)
	ctx := context.Background()
	siteID := uuid.New()

	// Create test user
	user, _ := domain.NewUser(siteID, "bench@example.com", "encrypted_password_hash_that_is_at_least_sixty_characters_long")
	repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.FindByEmail(ctx, siteID, "bench@example.com")
	}
}
