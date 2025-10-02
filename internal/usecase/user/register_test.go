package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// MockUserRepository for testing
type MockUserRepository struct {
	users       []*domain.User
	createError error
	findError   error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createError != nil {
		return m.createError
	}
	m.users = append(m.users, user)
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*domain.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}

	for _, user := range m.users {
		if user.SiteID == siteID && user.Email == email {
			return user, nil
		}
	}
	return nil, nil // User not found
}

func (m *MockUserRepository) Reset() {
	m.users = nil
	m.createError = nil
	m.findError = nil
}

// Test suite for RegisterUsecase
func TestRegisterUsecase_Success(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := RegisterInput{
		SiteID:   uuid.New(),
		Email:    "test@example.com",
		Password: "securepass123",
	}

	// Execute
	user, err := registerUC.Execute(ctx, input)

	// Validate - These assertions will fail until T013 is implemented
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, input.Email, user.Email)
	assert.Equal(t, input.SiteID, user.SiteID)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.PasswordHash)
	assert.True(t, len(user.PasswordHash) >= 60) // bcrypt hash length
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestRegisterUsecase_DuplicateEmail(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	siteID := uuid.New()
	email := "duplicate@example.com"

	// Pre-create a user with the same email
	existingUser := &domain.User{
		ID:           uuid.New(),
		SiteID:       siteID,
		Email:        email,
		PasswordHash: "existing-hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	mockRepo.users = append(mockRepo.users, existingUser)

	ctx := context.Background()
	input := RegisterInput{
		SiteID:   siteID,
		Email:    email,
		Password: "newpassword123",
	}

	// Execute
	user, err := registerUC.Execute(ctx, input)

	// Validate - Should fail with duplicate email error
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email already exists") // Will be implemented in T013
}

func TestRegisterUsecase_SameEmailDifferentSites(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	siteA := uuid.New()
	siteB := uuid.New()
	email := "user@example.com"

	// Create user in site A
	existingUser := &domain.User{
		ID:           uuid.New(),
		SiteID:       siteA,
		Email:        email,
		PasswordHash: "existing-hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	mockRepo.users = append(mockRepo.users, existingUser)

	ctx := context.Background()
	input := RegisterInput{
		SiteID:   siteB,
		Email:    email,
		Password: "newpassword123",
	}

	// Execute - Should succeed as it's a different site
	user, err := registerUC.Execute(ctx, input)

	// Validate
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, siteB, user.SiteID)
	assert.Equal(t, email, user.Email)
}

func TestRegisterUsecase_ContextCancellation(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	input := RegisterInput{
		SiteID:   uuid.New(),
		Email:    "test@example.com",
		Password: "securepass123",
	}

	// Execute
	user, err := registerUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, user)
	// Should handle context cancellation properly
}

func TestRegisterUsecase_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{
		createError: errors.New("database connection failed"),
	}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()
	input := RegisterInput{
		SiteID:   uuid.New(),
		Email:    "test@example.com",
		Password: "securepass123",
	}

	// Execute
	user, err := registerUC.Execute(ctx, input)

	// Validate
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestRegisterUsecase_InvalidInput(t *testing.T) {
	// Setup
	mockRepo := &MockUserRepository{}
	registerUC := NewRegisterUsecase(mockRepo, "test-secret-key-at-least-32-characters")
	defer mockRepo.Reset()

	ctx := context.Background()

	tests := []struct {
		name  string
		input RegisterInput
	}{
		{
			name: "empty site ID",
			input: RegisterInput{
				SiteID:   uuid.Nil,
				Email:    "test@example.com",
				Password: "securepass123",
			},
		},
		{
			name: "empty email",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "",
				Password: "securepass123",
			},
		},
		{
			name: "invalid email format",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "invalid-email",
				Password: "securepass123",
			},
		},
		{
			name: "password too short",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "short",
			},
		},
		{
			name: "empty password",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			user, err := registerUC.Execute(ctx, tt.input)

			// Validate
			assert.Error(t, err)
			assert.Nil(t, user)
		})
	}
}
