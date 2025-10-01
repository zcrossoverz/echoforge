// internal/usecase/user_usecase.go
package usecase

import (
	"context"
	"errors"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/pkg/auth"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// Register: Command - Create if not exists (business rule: unique email)
func (uc *UserUsecase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := domain.NewUser(email, password)
	if err != nil {
		return nil, err // Validation/hash error from domain
	}

	// Check exists (idempotent business rule)
	existing, err := uc.repo.FindByEmail(email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err // Propagate DB errors
	}
	if existing != nil {
		return nil, domain.ErrUserExists
	}

	// Create (transactional in real impl)
	if err := uc.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login: Query - Verify & gen token
func (uc *UserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.repo.FindByEmail(email)
	if err != nil {
		return "", err // Includes ErrUserNotFound
	}

	if !user.CheckPassword(password) {
		return "", domain.ErrWrongPassword
	}

	// Gen JWT (stub - full in pkg/auth với claims: userID, role, exp)
	token, err := auth.GenerateJWT(user.ID.String(), user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}
