package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

func TestNewUser_Valid(t *testing.T) {
	email := "user@example.com"
	password := "password123"
	user, err := domain.NewUser(email, password)
	assert.NoError(t, err)
	assert.NotNil(t, user) // Explicit nil check trước access
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.ID)          // UUID generated (non-zero)
	assert.Len(t, user.PasswordHash, 60) // Bcrypt default length
	assert.Equal(t, "user", user.Role)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second) // Add time import nếu cần
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := domain.NewUser("invalid-email", "pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Email") // Validator msg
}

func TestUser_CheckPassword(t *testing.T) {
	user, _ := domain.NewUser("test@example.com", "password123")
	assert.True(t, user.CheckPassword("password123"))
	assert.False(t, user.CheckPassword("wrong"))
}

// Stub repo test: Giữ nguyên
func TestUserRepository_Create(t *testing.T) {
	// Expand ở Task 1.3
}
