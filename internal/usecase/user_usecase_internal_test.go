package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// MockUserRepository for internal package testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// TestUserUseCase_InternalPackageTest tests UserUseCase functionality from within the usecase package
func TestUserUseCase_InternalPackageTest(t *testing.T) {
	mockRepo := new(MockUserRepository)
	useCase := NewUserUseCase(mockRepo)

	t.Run("CreateUser - Success", func(t *testing.T) {
		email := "test@example.com"
		passwordHash := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u" // bcrypt hash (60 chars)

		// Mock repository calls
		mockRepo.On("FindByEmail", mock.Anything, email).Return((*domain.User)(nil), nil).Once()
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		user, err := useCase.CreateUser(context.Background(), email, passwordHash)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, passwordHash, user.PasswordHash)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateUser - User already exists", func(t *testing.T) {
		email := "existing@example.com"
		passwordHash := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u" // bcrypt hash (60 chars)

		existingUser := &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u", // bcrypt hash (60 chars)
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockRepo.On("FindByEmail", mock.Anything, email).Return(existingUser, nil).Once()

		user, err := useCase.CreateUser(context.Background(), email, passwordHash)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetUserByEmail - Success", func(t *testing.T) {
		email := "test@example.com"
		expectedUser := &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u", // bcrypt hash (60 chars)
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockRepo.On("FindByEmail", mock.Anything, email).Return(expectedUser, nil).Once()

		user, err := useCase.GetUserByEmail(context.Background(), email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("IsEmailAvailable - Available", func(t *testing.T) {
		email := "available@example.com"

		mockRepo.On("FindByEmail", mock.Anything, email).Return((*domain.User)(nil), nil).Once()

		available, err := useCase.IsEmailAvailable(context.Background(), email)

		assert.NoError(t, err)
		assert.True(t, available)
		mockRepo.AssertExpectations(t)
	})

	t.Run("IsEmailAvailable - Not available", func(t *testing.T) {
		email := "taken@example.com"
		existingUser := &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewfyAdEeLsrCxO7u", // bcrypt hash (60 chars)
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockRepo.On("FindByEmail", mock.Anything, email).Return(existingUser, nil).Once()

		available, err := useCase.IsEmailAvailable(context.Background(), email)

		assert.NoError(t, err)
		assert.False(t, available)
		mockRepo.AssertExpectations(t)
	})
}

// TestNewUserUseCase tests the constructor
func TestNewUserUseCase(t *testing.T) {
	mockRepo := new(MockUserRepository)
	useCase := NewUserUseCase(mockRepo)

	assert.NotNil(t, useCase)
}
