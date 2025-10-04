package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
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

func TestNewUserUseCase(t *testing.T) {
	mockRepo := &MockUserRepository{}
	useCase := usecase.NewUserUseCase(mockRepo)

	assert.NotNil(t, useCase)
}

func TestUserUseCase_CreateUser(t *testing.T) {
	// Create a valid bcrypt hash for testing
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		email        string
		passwordHash string
		setupMock    func(*MockUserRepository)
		expectErr    bool
		errMsg       string
	}{
		{
			name:         "Successful user creation",
			email:        "test@example.com",
			passwordHash: string(validPasswordHash),
			setupMock: func(m *MockUserRepository) {
				// Return nil for FindByEmail (user doesn't exist)
				m.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), nil).Once()
				// Return nil for Create (success)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name:         "User already exists",
			email:        "existing@example.com",
			passwordHash: string(validPasswordHash),
			setupMock: func(m *MockUserRepository) {
				existingUser := &domain.User{
					ID:           uuid.New(),
					Email:        "existing@example.com",
					PasswordHash: string(validPasswordHash),
				}
				m.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil).Once()
			},
			expectErr: true,
			errMsg:    "already exists",
		},
		{
			name:         "Empty email",
			email:        "",
			passwordHash: string(validPasswordHash),
			setupMock:    func(m *MockUserRepository) {},
			expectErr:    true,
			errMsg:       "cannot be empty",
		},
		{
			name:         "Empty password hash",
			email:        "test@example.com",
			passwordHash: "",
			setupMock:    func(m *MockUserRepository) {},
			expectErr:    true,
			errMsg:       "cannot be empty",
		},
		{
			name:         "Repository FindByEmail error",
			email:        "test@example.com",
			passwordHash: string(validPasswordHash),
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), errors.New("database error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to check existing user",
		},
		{
			name:         "Repository Create error",
			email:        "test@example.com",
			passwordHash: string(validPasswordHash),
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), nil).Once()
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("create failed")).Once()
			},
			expectErr: true,
			errMsg:    "failed to create user",
		},
		{
			name:         "Invalid email format",
			email:        "invalid-email",
			passwordHash: string(validPasswordHash),
			setupMock:    func(m *MockUserRepository) {},
			expectErr:    true,
			errMsg:       "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			useCase := usecase.NewUserUseCase(mockRepo)
			ctx := context.Background()

			user, err := useCase.CreateUser(ctx, tt.email, tt.passwordHash)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.passwordHash, user.PasswordHash)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_CreateUser_ContextCancellation(t *testing.T) {
	mockRepo := &MockUserRepository{}
	useCase := usecase.NewUserUseCase(mockRepo)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user, err := useCase.CreateUser(ctx, "test@example.com", string(validPasswordHash))

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestUserUseCase_GetUserByEmail(t *testing.T) {
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(validPasswordHash),
	}

	tests := []struct {
		name      string
		email     string
		setupMock func(*MockUserRepository)
		expectErr bool
		errMsg    string
		expected  *domain.User
	}{
		{
			name:  "User found",
			email: "test@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil).Once()
			},
			expectErr: false,
			expected:  existingUser,
		},
		{
			name:  "User not found",
			email: "notfound@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "notfound@example.com").Return((*domain.User)(nil), nil).Once()
			},
			expectErr: false,
			expected:  nil,
		},
		{
			name:      "Empty email",
			email:     "",
			setupMock: func(m *MockUserRepository) {},
			expectErr: true,
			errMsg:    "cannot be empty",
		},
		{
			name:  "Repository error",
			email: "test@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), errors.New("database error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to find user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			useCase := usecase.NewUserUseCase(mockRepo)
			ctx := context.Background()

			user, err := useCase.GetUserByEmail(ctx, tt.email)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.expected != nil {
					assert.Equal(t, tt.expected.ID, user.ID)
					assert.Equal(t, tt.expected.Email, user.Email)
				} else {
					assert.Nil(t, user)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_GetUserByEmail_ContextCancellation(t *testing.T) {
	mockRepo := &MockUserRepository{}
	useCase := usecase.NewUserUseCase(mockRepo)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	user, err := useCase.GetUserByEmail(ctx, "test@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestUserUseCase_IsEmailAvailable(t *testing.T) {
	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "existing@example.com",
		PasswordHash: string(validPasswordHash),
	}

	tests := []struct {
		name              string
		email             string
		setupMock         func(*MockUserRepository)
		expectErr         bool
		errMsg            string
		expectedAvailable bool
	}{
		{
			name:  "Email available",
			email: "available@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "available@example.com").Return((*domain.User)(nil), nil).Once()
			},
			expectErr:         false,
			expectedAvailable: true,
		},
		{
			name:  "Email not available",
			email: "existing@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil).Once()
			},
			expectErr:         false,
			expectedAvailable: false,
		},
		{
			name:      "Empty email",
			email:     "",
			setupMock: func(m *MockUserRepository) {},
			expectErr: true,
			errMsg:    "cannot be empty",
		},
		{
			name:  "Repository error",
			email: "test@example.com",
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), errors.New("database error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to check email availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			useCase := usecase.NewUserUseCase(mockRepo)
			ctx := context.Background()

			available, err := useCase.IsEmailAvailable(ctx, tt.email)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAvailable, available)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_IsEmailAvailable_ContextCancellation(t *testing.T) {
	mockRepo := &MockUserRepository{}
	useCase := usecase.NewUserUseCase(mockRepo)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	available, err := useCase.IsEmailAvailable(ctx, "test@example.com")

	assert.Error(t, err)
	assert.False(t, available)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestUserUseCase_EdgeCases(t *testing.T) {
	t.Run("Create user with domain validation failure", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		useCase := usecase.NewUserUseCase(mockRepo)
		ctx := context.Background()

		// Try to create user with invalid email that domain validation will catch
		user, err := useCase.CreateUser(ctx, "invalid-email", "short") // Too short password hash

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to create user entity")
	})

	t.Run("Create user with very long valid email", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		mockRepo.On("FindByEmail", mock.Anything, mock.AnythingOfType("string")).Return((*domain.User)(nil), nil).Once()
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		useCase := usecase.NewUserUseCase(mockRepo)
		ctx := context.Background()

		validPasswordHash, err := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
		assert.NoError(t, err)

		// Create a long but valid email (within 255 char limit)
		longEmail := string(make([]byte, 240)) + "@a.com"
		for i := range longEmail[:240] {
			longEmail = longEmail[:i] + "a" + longEmail[i+1:]
		}

		user, err := useCase.CreateUser(ctx, longEmail, string(validPasswordHash))

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, longEmail, user.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetUserByEmail with nil repository response", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), nil).Once()

		useCase := usecase.NewUserUseCase(mockRepo)
		ctx := context.Background()

		user, err := useCase.GetUserByEmail(ctx, "test@example.com")

		assert.NoError(t, err)
		assert.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}
