package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

func TestNewUser_Success(t *testing.T) {
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	user, err := domain.NewUser(email, passwordHash)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestNewUser_ValidationErrors(t *testing.T) {
	validEmail := "valid@example.com"
	validPasswordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	tests := []struct {
		name         string
		email        string
		passwordHash string
		expectErr    string
	}{
		{
			name:         "empty email",
			email:        "",
			passwordHash: validPasswordHash,
			expectErr:    "Email",
		},
		{
			name:         "invalid email format",
			email:        "invalid-email",
			passwordHash: validPasswordHash,
			expectErr:    "invalid email format",
		},
		{
			name:         "email too long",
			email:        "a_very_long_email_address_that_definitely_exceeds_the_maximum_allowed_length_of_two_hundred_fifty_five_characters_for_email_validation_purposes_in_the_user_domain_entity_which_should_cause_this_test_to_fail_properly_and_demonstrate_the_validation_working_correctly_with_extra_characters_to_push_it_over_the_limit@example.com",
			passwordHash: validPasswordHash,
			expectErr:    "email exceeds maximum length",
		},
		{
			name:         "empty password hash",
			email:        validEmail,
			passwordHash: "",
			expectErr:    "PasswordHash",
		},
		{
			name:         "password hash too short",
			email:        validEmail,
			passwordHash: "short_hash",
			expectErr:    "password hash too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.passwordHash)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name      string
		user      func() *domain.User
		expectErr string
	}{
		{
			name: "valid user",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					Email:        "valid@example.com",
					PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "",
		},
		{
			name: "nil ID",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.Nil,
					Email:        "valid@example.com",
					PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "ID",
		},
		{
			name: "invalid email format",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					Email:        "invalid-email",
					PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "invalid email format",
		},
		{
			name: "password hash too short",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					Email:        "valid@example.com",
					PasswordHash: "short",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "password hash too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.user()
			err := user.Validate()

			if tt.expectErr == "" {
				assert.NoError(t, err)
				assert.True(t, user.IsValid())
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				assert.False(t, user.IsValid())
			}
		})
	}
}

func TestCloneAndExtendModel(t *testing.T) {
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Each clone instance has separate database, so email uniqueness is enforced globally within clone
	user, err := domain.NewUser(email, passwordHash)
	require.NoError(t, err)

	// Verify user created without site context
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
}
