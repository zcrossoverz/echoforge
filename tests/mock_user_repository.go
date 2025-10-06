package tests

import (
	"context"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// MockUserRepository implements domain.UserRepository for testing
type MockUserRepository struct {
	users map[string]*domain.User // key: email
	calls map[string]int          // method call counts
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
		calls: make(map[string]int),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.calls["Create"]++

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if user == nil {
		return domain.ErrRepositoryFailure
	}

	if err := user.Validate(); err != nil {
		return err
	}

	key := user.Email
	if _, exists := m.users[key]; exists {
		return domain.ErrUserAlreadyExists
	}

	// Create a copy to avoid mutations
	userCopy := &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	m.users[key] = userCopy
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.calls["FindByEmail"]++

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if email == "" {
		return nil, domain.ErrRepositoryFailure
	}

	key := email
	user, exists := m.users[key]
	if !exists {
		return nil, nil
	}

	// Return a copy to avoid mutations
	userCopy := &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	return userCopy, nil
}

// Test helpers
func (m *MockUserRepository) CallCount(method string) int {
	return m.calls[method]
}

func (m *MockUserRepository) Reset() {
	m.users = make(map[string]*domain.User)
	m.calls = make(map[string]int)
}

// GetAllUsers returns all stored users (for testing purposes)
func (m *MockUserRepository) GetAllUsers() map[string]*domain.User {
	result := make(map[string]*domain.User)
	for key, user := range m.users {
		result[key] = &domain.User{
			ID:           user.ID,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}
	}
	return result
}

// CountUsers returns the total number of users stored
func (m *MockUserRepository) CountUsers() int {
	return len(m.users)
}

// DeleteUser removes a user (for test cleanup)
func (m *MockUserRepository) DeleteUser(email string) bool {
	key := email
	if _, exists := m.users[key]; exists {
		delete(m.users, key)
		return true
	}
	return false
}

// CorruptUser simulates data corruption for error testing
func (m *MockUserRepository) CorruptUser(email string) {
	key := email
	if user, exists := m.users[key]; exists {
		// Corrupt the user data
		user.Email = "corrupted"
		user.PasswordHash = "short" // Invalid hash length
	}
}
