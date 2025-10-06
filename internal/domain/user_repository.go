package domain

import (
	"context"
)

// UserRepository defines the persistence contract for User entities
// This interface follows hexagonal architecture principles - the domain defines the contract,
// and infrastructure adapters implement it
type UserRepository interface {
	// Create persists a new user entity
	// Returns ErrUserAlreadyExists if email already exists
	Create(ctx context.Context, user *User) error

	// FindByEmail retrieves a user by email address
	// Returns nil, nil if user not found (not an error condition)
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByID retrieves a user by their UUID
	// Returns nil, nil if user not found (not an error condition)
	FindByID(ctx context.Context, id string) (*User, error)

	// Update modifies an existing user entity
	// Updates the UpdatedAt timestamp automatically
	Update(ctx context.Context, user *User) error

	// Delete removes a user entity (soft delete recommended in production)
	Delete(ctx context.Context, id string) error

	// ExistsByEmail checks if a user with the given email exists
	// More efficient than FindByEmail when only existence check is needed
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
