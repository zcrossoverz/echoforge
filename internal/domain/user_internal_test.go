package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUser_InternalPackageTest tests User functionality from within the domain package
func TestUser_InternalPackageTest(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid user creation",
			email:    "test@example.com",
			password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u", // bcrypt hash (60 chars)
			wantErr:  false,
		},
		{
			name:     "Invalid email",
			email:    "invalid-email",
			password: "validhashedpassword123",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			email:    "test@example.com",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.password, user.PasswordHash)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
			}
		})
	}
}

// TestUser_ValidationMethods tests the validation methods
func TestUser_ValidationMethods(t *testing.T) {
	user := &User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u", // bcrypt hash (60 chars)
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test IsValid
	assert.True(t, user.IsValid())

	// Test Validate
	assert.NoError(t, user.Validate())

	// Test with invalid email
	user.Email = "invalid-email"
	assert.False(t, user.IsValid())
	assert.Error(t, user.Validate())
}
