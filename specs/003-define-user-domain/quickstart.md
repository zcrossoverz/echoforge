# Quickstart Guide: User Domain Entity

**Feature**: User Domain Entity and Repository  
**Date**: 2025-10-02  
**Prerequisites**: Go 1.25+, PostgreSQL 16+

## Overview
This guide demonstrates how to create, validate, and persist User entities with multi-tenant isolation. Follow these steps to implement and test the User domain functionality.

## Step 1: Setup Dependencies

Ensure your `go.mod` includes required dependencies:
```bash
# Navigate to project root
cd /path/to/echoforge

# Verify dependencies (should already exist)
go mod tidy
```

Required packages:
- `github.com/google/uuid` - UUID generation
- `gorm.io/gorm` - ORM for persistence  
- `github.com/stretchr/testify` - Testing framework
- `golang.org/x/crypto/bcrypt` - Password hashing (for validation)

## Step 2: Implement User Entity

Create `internal/domain/user.go`:

```go
package domain

import (
    "errors"
    "fmt"
    "regexp"
    "time"
    
    "github.com/google/uuid"
)

// User represents a registered user within a site
type User struct {
    ID           uuid.UUID `json:"id"`
    SiteID       uuid.UUID `json:"site_id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validation constants
const (
    MaxEmailLength    = 255
    MinPasswordLength = 60 // bcrypt hash length
)

// Domain errors
var (
    ErrInvalidEmail        = errors.New("invalid email format")
    ErrEmailTooLong       = errors.New("email exceeds maximum length")
    ErrPasswordHashTooShort = errors.New("password hash too short")
    ErrRequiredField      = errors.New("required field is empty")
)

// NewUser creates a new User entity with validation
func NewUser(siteID uuid.UUID, email, passwordHash string) (*User, error) {
    now := time.Now()
    
    user := &User{
        ID:           uuid.New(),
        SiteID:       siteID,
        Email:        email,
        PasswordHash: passwordHash,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    
    if err := user.Validate(); err != nil {
        return nil, err
    }
    
    return user, nil
}

// Validate performs business rule validation
func (u *User) Validate() error {
    // Required fields
    if u.ID == uuid.Nil {
        return fmt.Errorf("ID: %w", ErrRequiredField)
    }
    if u.SiteID == uuid.Nil {
        return fmt.Errorf("SiteID: %w", ErrRequiredField)
    }
    if u.Email == "" {
        return fmt.Errorf("Email: %w", ErrRequiredField)
    }
    if u.PasswordHash == "" {
        return fmt.Errorf("PasswordHash: %w", ErrRequiredField)
    }
    
    // Email validation
    if len(u.Email) > MaxEmailLength {
        return fmt.Errorf("Email: %w (%d > %d)", ErrEmailTooLong, len(u.Email), MaxEmailLength)
    }
    if !emailRegex.MatchString(u.Email) {
        return fmt.Errorf("Email: %w", ErrInvalidEmail)
    }
    
    // Password hash validation
    if len(u.PasswordHash) < MinPasswordLength {
        return fmt.Errorf("PasswordHash: %w (%d < %d)", ErrPasswordHashTooShort, len(u.PasswordHash), MinPasswordLength)
    }
    
    return nil
}

// IsValid returns true if entity passes all validation rules
func (u *User) IsValid() bool {
    return u.Validate() == nil
}
```

## Step 3: Define Repository Interface

Add to `internal/domain/user.go`:

```go
import "context"

// UserRepository defines the persistence contract for User entities
type UserRepository interface {
    // Create persists a new user entity
    Create(ctx context.Context, user *User) error
    
    // FindByEmail retrieves a user by email within specific site
    FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error)
}

// Repository errors
var (
    ErrUserAlreadyExists = errors.New("user already exists with this email in site")
    ErrRepositoryFailure = errors.New("repository operation failed")
)
```

## Step 4: Create Unit Tests

Create `tests/user_domain_test.go`:

```go
package tests

import (
    "testing"
    "time"
    
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/zcrossoverz/echoforge/internal/domain"
)

func TestNewUser_Success(t *testing.T) {
    siteID := uuid.New()
    email := "user@example.com"
    passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"
    
    user, err := domain.NewUser(siteID, email, passwordHash)
    
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, user.ID)
    assert.Equal(t, siteID, user.SiteID)
    assert.Equal(t, email, user.Email)
    assert.Equal(t, passwordHash, user.PasswordHash)
    assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
    assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestUser_Validate(t *testing.T) {
    tests := []struct {
        name      string
        user      func() *domain.User
        expectErr string
    }{
        {
            name: "valid user",
            user: func() *domain.User {
                return &domain.User{
                    ID:           uuid.New(),
                    SiteID:       uuid.New(),
                    Email:        "valid@example.com",
                    PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
                    CreatedAt:    time.Now(),
                    UpdatedAt:    time.Now(),
                }
            },
            expectErr: "",
        },
        {
            name: "invalid email format",
            user: func() *domain.User {
                return &domain.User{
                    ID:           uuid.New(),
                    SiteID:       uuid.New(),
                    Email:        "invalid-email",
                    PasswordHash: "encrypted_password_hash_that_is_at_least_sixty_characters_long",
                    CreatedAt:    time.Now(),
                    UpdatedAt:    time.Now(),
                }
            },
            expectErr: "invalid email format",
        },
        {
            name: "password hash too short",
            user: func() *domain.User {
                return &domain.User{
                    ID:           uuid.New(),
                    SiteID:       uuid.New(),
                    Email:        "valid@example.com",
                    PasswordHash: "short",
                    CreatedAt:    time.Now(),
                    UpdatedAt:    time.Now(),
                }
            },
            expectErr: "password hash too short",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user := tt.user()
            err := user.Validate()
            
            if tt.expectErr == "" {
                assert.NoError(t, err)
                assert.True(t, user.IsValid())
            } else {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectErr)
                assert.False(t, user.IsValid())
            }
        })
    }
}
```

## Step 5: Run Tests

```bash
# Run domain tests
go test ./tests/user_domain_test.go -v

# Expected output:
# === RUN   TestNewUser_Success
# --- PASS: TestNewUser_Success (0.00s)
# === RUN   TestUser_Validate
# === RUN   TestUser_Validate/valid_user
# --- PASS: TestUser_Validate/valid_user (0.00s)
# === RUN   TestUser_Validate/invalid_email_format
# --- PASS: TestUser_Validate/invalid_email_format (0.00s)
# === RUN   TestUser_Validate/password_hash_too_short
# --- PASS: TestUser_Validate/password_hash_too_short (0.00s)
# --- PASS: TestUser_Validate (0.00s)
# PASS
```

## Step 6: Create Database Migration

Create `migrations/002_create_users_table.up.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT users_email_length CHECK (char_length(email) <= 255),
    CONSTRAINT users_password_hash_length CHECK (char_length(password_hash) >= 60),
    CONSTRAINT users_site_email_unique UNIQUE (site_id, email)
);

CREATE INDEX idx_users_site_id ON users(site_id);
CREATE INDEX idx_users_site_email ON users(site_id, email);
```

Create `migrations/002_create_users_table.down.sql`:

```sql
DROP TABLE IF EXISTS users;
```

## Step 7: Example Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    
    "github.com/zcrossoverz/echoforge/internal/domain"
)

func ExampleUserCreation() {
    // Generate site ID (would come from config in real app)
    siteID := uuid.New()
    
    // Hash password (would be done in auth layer)
    plainPassword := "user-password-123"
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    
    // Create user entity
    user, err := domain.NewUser(siteID, "user@example.com", string(hashedPassword))
    if err != nil {
        log.Fatal("Failed to create user:", err)
    }
    
    fmt.Printf("Created user: %+v\n", user)
    fmt.Printf("User is valid: %t\n", user.IsValid())
    
    // Would persist with repository:
    // err = userRepo.Create(context.Background(), user)
}
```

## Step 8: Multi-Tenant Validation

Test site isolation:

```go
func TestMultiTenantIsolation(t *testing.T) {
    siteA := uuid.New()
    siteB := uuid.New()
    email := "user@example.com"
    passwordHash := "encrypted_password_hash_that_is_at_least_sixty_characters_long"
    
    // Same email can exist in different sites
    userA, err := domain.NewUser(siteA, email, passwordHash)
    require.NoError(t, err)
    
    userB, err := domain.NewUser(siteB, email, passwordHash)
    require.NoError(t, err)
    
    // Different users despite same email
    assert.NotEqual(t, userA.ID, userB.ID)
    assert.NotEqual(t, userA.SiteID, userB.SiteID)
    assert.Equal(t, userA.Email, userB.Email)
}
```

## Next Steps

1. **Implement Repository**: Create GORM implementation in `adapters/persistence/user_repository.go`
2. **Add Use Cases**: Implement business logic in `internal/usecase/user_usecase.go`  
3. **Integration Tests**: Test repository with real database
4. **API Layer**: Add HTTP endpoints using Gin (future)

## Common Issues

### Password Hash Validation Fails
- Ensure password is hashed with bcrypt before creating User entity
- bcrypt produces 60-character hashes - verify hash length

### Email Validation Fails  
- Check email format matches RFC 5322 requirements
- Verify email length doesn't exceed 255 characters

### Site Isolation Issues
- Always include `site_id` in repository queries
- Test cross-site access prevention in integration tests

### UUID Generation
- Use `uuid.New()` for v4 UUIDs (random)
- Never use `uuid.Nil` for entity IDs

This quickstart provides a complete foundation for the User domain entity with proper validation, testing, and multi-tenant support.