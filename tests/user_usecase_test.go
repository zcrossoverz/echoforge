package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// MockUserRepo: Giữ nguyên (từ hướng dẫn trước)
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// Các test khác giữ nguyên: TestUserUsecase_Register_Success, Register_AlreadyExists, Login_WrongPassword, Login_NotFound

func TestUserUsecase_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	uc := usecase.NewUserUsecase(mockRepo)

	// Setup: Gen real hash từ domain (unpack explicit để fix compile)
	setupUser, err := domain.NewUser("login@example.com", "password")
	if err != nil {
		t.Fatal(err) // Halt nếu setup fail (TDD safety)
	}
	hashedPw := setupUser.PasswordHash // Now safe access

	// Mock user với real hash
	user := &domain.User{
		ID:           uuid.New(),
		Email:        "login@example.com",
		PasswordHash: hashedPw, // Real hash cho CheckPassword verify
		Role:         "user",
		CreatedAt:    time.Now(),
	}
	mockRepo.On("FindByEmail", "login@example.com").Return(user, nil)

	// Call usecase
	token, err := uc.Login(context.Background(), "login@example.com", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token) // Stub token check

	mockRepo.AssertExpectations(t)
}
