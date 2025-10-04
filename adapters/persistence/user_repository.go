package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// GormUser represents the database model for users
type GormUser struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash string    `gorm:"type:varchar(60);not null"`
	CreatedAt    int64     `gorm:"autoCreateTime"`
	UpdatedAt    int64     `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (GormUser) TableName() string {
	return "users"
}

// UserRepository implements the domain UserRepository interface using GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM-based user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create persists a new user entity
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if user == nil {
		return fmt.Errorf("user cannot be nil: %w", domain.ErrRepositoryFailure)
	}

	// Validate domain entity
	if err := user.Validate(); err != nil {
		return err
	}

	// Convert domain entity to GORM model
	gormUser := &GormUser{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Unix(),
		UpdatedAt:    user.UpdatedAt.Unix(),
	}

	// Execute database operation with context
	result := r.db.WithContext(ctx).Create(gormUser)
	if result.Error != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(result.Error) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", domain.ErrRepositoryFailure)
	}

	return nil
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if email == "" {
		return nil, fmt.Errorf("email cannot be empty: %w", domain.ErrRepositoryFailure)
	}

	var gormUser GormUser
	result := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&gormUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find user: %w", domain.ErrRepositoryFailure)
	}

	// Convert GORM model to domain entity
	domainUser := &domain.User{
		ID:           gormUser.ID,
		Email:        gormUser.Email,
		PasswordHash: gormUser.PasswordHash,
		CreatedAt:    unixToTime(gormUser.CreatedAt),
		UpdatedAt:    unixToTime(gormUser.UpdatedAt),
	}

	// Validate the retrieved entity
	if err := domainUser.Validate(); err != nil {
		return nil, fmt.Errorf("retrieved user is invalid: %w", err)
	}

	return domainUser, nil
}

// FindByID retrieves a user by their UUID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if id == "" {
		return nil, fmt.Errorf("id cannot be empty: %w", domain.ErrRepositoryFailure)
	}

	// Parse UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var gormUser GormUser
	result := r.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&gormUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find user: %w", domain.ErrRepositoryFailure)
	}

	// Convert GORM model to domain entity
	domainUser := &domain.User{
		ID:           gormUser.ID,
		Email:        gormUser.Email,
		PasswordHash: gormUser.PasswordHash,
		CreatedAt:    unixToTime(gormUser.CreatedAt),
		UpdatedAt:    unixToTime(gormUser.UpdatedAt),
	}

	// Validate the retrieved entity
	if err := domainUser.Validate(); err != nil {
		return nil, fmt.Errorf("retrieved user is invalid: %w", err)
	}

	return domainUser, nil
}

// Update modifies an existing user entity
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if user == nil {
		return fmt.Errorf("user cannot be nil: %w", domain.ErrRepositoryFailure)
	}

	// Validate domain entity
	if err := user.Validate(); err != nil {
		return err
	}

	// Convert domain entity to GORM model
	gormUser := &GormUser{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Unix(),
		UpdatedAt:    time.Now().Unix(), // Update timestamp
	}

	// Execute database operation with context
	result := r.db.WithContext(ctx).Model(&GormUser{}).Where("id = ?", user.ID).Updates(gormUser)
	if result.Error != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(result.Error) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", domain.ErrRepositoryFailure)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrRepositoryFailure)
	}

	// Update the domain entity with new timestamp
	user.UpdatedAt = unixToTime(gormUser.UpdatedAt)

	return nil
}

// Delete removes a user entity
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if id == "" {
		return fmt.Errorf("id cannot be empty: %w", domain.ErrRepositoryFailure)
	}

	// Parse UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Execute database operation with context
	result := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&GormUser{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", domain.ErrRepositoryFailure)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrRepositoryFailure)
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if email == "" {
		return false, fmt.Errorf("email cannot be empty: %w", domain.ErrRepositoryFailure)
	}

	var count int64
	result := r.db.WithContext(ctx).Model(&GormUser{}).Where("email = ?", email).Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check email existence: %w", domain.ErrRepositoryFailure)
	}

	return count > 0, nil
}

// Helper functions

// isDuplicateKeyError checks if the error is a unique constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	// PostgreSQL unique constraint violation
	return contains(errorStr, "duplicate key value violates unique constraint") ||
		contains(errorStr, "UNIQUE constraint failed") ||
		contains(errorStr, "Error 1062") // MySQL duplicate entry
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring performs case-insensitive substring search
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	sLower := toLowerCase(s)
	substrLower := toLowerCase(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLowerCase converts a string to lowercase
func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + ('a' - 'A')
		} else {
			result[i] = b
		}
	}
	return string(result)
}

// unixToTime converts Unix timestamp to time.Time
func unixToTime(unix int64) time.Time {
	return time.Unix(unix, 0).UTC()
}
