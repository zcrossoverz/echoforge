// Package contracts defines the updated use case interfaces for clone-and-extend model
package contracts

import (
	"context"
	"time"
)

// RegisterInput represents the updated input for user registration (site_id removed)
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// LoginInput represents the updated input for user login (site_id removed)
type LoginInput struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,max=128"`
}

// AuthenticationResult represents the result of successful authentication
type AuthenticationResult struct {
	User      *User     `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterUsecase defines the interface for user registration business logic
type RegisterUsecase interface {
	Execute(ctx context.Context, input RegisterInput) (*User, error)
}

// LoginUsecase defines the interface for user login business logic
type LoginUsecase interface {
	Execute(ctx context.Context, input LoginInput) (*AuthenticationResult, error)
}

// Use case validation interface
type InputValidator interface {
	Validate(input interface{}) error
}
