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
	siteID := uuid.New()
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	user, err := domain.NewUser(siteID, email, passwordHash)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, siteID, user.SiteID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestNewUser_ValidationErrors(t *testing.T) {
	siteID := uuid.New()
	validEmail := "valid@example.com"
	validPasswordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	tests := []struct {
		name         string
		siteID       uuid.UUID
		email        string
		passwordHash string
		expectErr    string
	}{
		{
			name:         "empty site ID",
			siteID:       uuid.Nil,
			email:        validEmail,
			passwordHash: validPasswordHash,
			expectErr:    "SiteID",
		},
		{
			name:         "empty email",
			siteID:       siteID,
			email:        "",
			passwordHash: validPasswordHash,
			expectErr:    "Email",
		},
		{
			name:         "invalid email format",
			siteID:       siteID,
			email:        "invalid-email",
			passwordHash: validPasswordHash,
			expectErr:    "invalid email format",
		},
		{
			name:         "email too long",
			siteID:       siteID,
			email:        "a_very_long_email_address_that_definitely_exceeds_the_maximum_allowed_length_of_two_hundred_fifty_five_characters_for_email_validation_purposes_in_the_user_domain_entity_which_should_cause_this_test_to_fail_properly_and_demonstrate_the_validation_working_correctly_with_extra_characters_to_push_it_over_the_limit@example.com",
			passwordHash: validPasswordHash,
			expectErr:    "email exceeds maximum length",
		},
		{
			name:         "empty password hash",
			siteID:       siteID,
			email:        validEmail,
			passwordHash: "",
			expectErr:    "PasswordHash",
		},
		{
			name:         "password hash too short",
			siteID:       siteID,
			email:        validEmail,
			passwordHash: "short_hash",
			expectErr:    "password hash too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.siteID, tt.email, tt.passwordHash)

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
					SiteID:       uuid.New(),
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
					SiteID:       uuid.New(),
					Email:        "valid@example.com",
					PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "ID",
		},
		{
			name: "nil site ID",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					SiteID:       uuid.Nil,
					Email:        "valid@example.com",
					PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			},
			expectErr: "SiteID",
		},
		{
			name: "invalid email format",
			user: func() *domain.User {
				return &domain.User{
					ID:           uuid.New(),
					SiteID:       uuid.New(),
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
					SiteID:       uuid.New(),
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

func TestMultiTenantIsolation(t *testing.T) {
	siteA := uuid.New()
	siteB := uuid.New()
	email := "user@example.com"
	passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"

	// Same email can exist in different sites
	userA, err := domain.NewUser(siteA, email, passwordHash)
	require.NoError(t, err)

	userB, err := domain.NewUser(siteB, email, passwordHash)
	require.NoError(t, err)

	// Different users despite same email
	assert.NotEqual(t, userA.ID, userB.ID)
	assert.NotEqual(t, userA.SiteID, userB.SiteID)
	assert.Equal(t, userA.Email, userB.Email)
	assert.Equal(t, userA.PasswordHash, userB.PasswordHash)
}
