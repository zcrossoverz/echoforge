package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

func TestNewUser(t *testing.T) {
	// Create a valid bcrypt hash for testing
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		email        string
		passwordHash string
		expectErr    bool
	}{
		{
			name:         "Valid user creation",
			email:        "test@example.com",
			passwordHash: string(validPasswordHash),
			expectErr:    false,
		},
		{
			name:         "Empty email",
			email:        "",
			passwordHash: string(validPasswordHash),
			expectErr:    true,
		},
		{
			name:         "Invalid email format",
			email:        "invalid-email",
			passwordHash: string(validPasswordHash),
			expectErr:    true,
		},
		{
			name:         "Email too long",
			email:        string(make([]byte, 256)) + "@example.com", // Creates email > 255 chars
			passwordHash: string(validPasswordHash),
			expectErr:    true,
		},
		{
			name:         "Empty password hash",
			email:        "test@example.com",
			passwordHash: "",
			expectErr:    true,
		},
		{
			name:         "Short password hash",
			email:        "test@example.com",
			passwordHash: "tooshort",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.passwordHash)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.passwordHash, user.PasswordHash)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
			}
		})
	}
}

func TestUser_Validate(t *testing.T) {
	// Create a valid bcrypt hash for testing
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		setupUser func() *domain.User
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid user",
			setupUser: func() *domain.User {
				user, _ := domain.NewUser("test@example.com", string(validPasswordHash))
				return user
			},
			expectErr: false,
		},
		{
			name: "Empty email",
			setupUser: func() *domain.User {
				user := &domain.User{
					ID:           uuid.New(),
					Email:        "",
					PasswordHash: string(validPasswordHash),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				return user
			},
			expectErr: true,
			errMsg:    "Email",
		},
		{
			name: "Invalid email format",
			setupUser: func() *domain.User {
				user := &domain.User{
					ID:           uuid.New(),
					Email:        "invalid-email",
					PasswordHash: string(validPasswordHash),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				return user
			},
			expectErr: true,
			errMsg:    "Email",
		},
		{
			name: "Email too long",
			setupUser: func() *domain.User {
				longEmail := string(make([]byte, 256)) + "@example.com" // Creates email > 255 chars
				user := &domain.User{
					ID:           uuid.New(),
					Email:        longEmail,
					PasswordHash: string(validPasswordHash),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				return user
			},
			expectErr: true,
			errMsg:    "Email",
		},
		{
			name: "Empty password hash",
			setupUser: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: "",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: true,
			errMsg:    "PasswordHash",
		},
		{
			name: "Short password hash",
			setupUser: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: "tooshort",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: true,
			errMsg:    "PasswordHash",
		},
		{
			name: "Nil UUID",
			setupUser: func() *domain.User {
				return &domain.User{
					ID:           uuid.Nil,
					Email:        "test@example.com",
					PasswordHash: string(validPasswordHash),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: true,
			errMsg:    "ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setupUser()
			err := user.Validate()

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_IsValid(t *testing.T) {
	// Create a valid bcrypt hash for testing
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	validUser, err := domain.NewUser("test@example.com", string(validPasswordHash))
	assert.NoError(t, err)

	invalidUser := &domain.User{
		ID:           uuid.Nil,
		Email:        "",
		PasswordHash: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.True(t, validUser.IsValid())
	assert.False(t, invalidUser.IsValid())
}

func TestUser_EdgeCases(t *testing.T) {
	// Create a valid bcrypt hash for testing
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	t.Run("Email at maximum length", func(t *testing.T) {
		// Create email exactly at 255 character limit
		baseEmail := "@email.com"
		prefix := string(make([]byte, 255-len(baseEmail)))
		for i := range prefix {
			prefix = prefix[:i] + "a" + prefix[i+1:]
		}
		longEmail := prefix + baseEmail // 255 chars total
		user, err := domain.NewUser(longEmail, string(validPasswordHash))
		assert.NoError(t, err)
		assert.Equal(t, longEmail, user.Email)
	})

	t.Run("Password hash at minimum length", func(t *testing.T) {
		// bcrypt hash should be 60 characters
		minHash := string(make([]byte, 60))
		user, err := domain.NewUser("test@example.com", minHash)
		assert.NoError(t, err)
		assert.Equal(t, minHash, user.PasswordHash)
	})

	t.Run("User timestamps are set", func(t *testing.T) {
		before := time.Now()
		user, err := domain.NewUser("test@example.com", string(validPasswordHash))
		after := time.Now()

		assert.NoError(t, err)
		assert.True(t, user.CreatedAt.After(before) || user.CreatedAt.Equal(before))
		assert.True(t, user.CreatedAt.Before(after) || user.CreatedAt.Equal(after))
		assert.True(t, user.UpdatedAt.After(before) || user.UpdatedAt.Equal(before))
		assert.True(t, user.UpdatedAt.Before(after) || user.UpdatedAt.Equal(after))
	})

	t.Run("UUID is generated", func(t *testing.T) {
		user, err := domain.NewUser("test@example.com", string(validPasswordHash))
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)

		// Generate another user to ensure different UUIDs
		user2, err := domain.NewUser("test2@example.com", string(validPasswordHash))
		assert.NoError(t, err)
		assert.NotEqual(t, user.ID, user2.ID)
	})
}
