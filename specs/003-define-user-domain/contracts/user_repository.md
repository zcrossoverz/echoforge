# Repository Contract: UserRepository

**Interface**: `UserRepository`  
**Package**: `internal/domain`  
**Date**: 2025-10-02

## Interface Definition

```go
package domain

import (
    "context"
    "github.com/google/uuid"
)

// UserRepository defines the persistence contract for User entities
type UserRepository interface {
    // Create persists a new user entity
    // Returns error if user already exists (email uniqueness within site)
    // Returns validation error if user data is invalid
    Create(ctx context.Context, user *User) error
    
    // FindByEmail retrieves a user by email within specific site
    // Returns nil, nil if user not found (not an error condition)
    // Returns error only for infrastructure failures
    FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error)
}
```

## Method Contracts

### Create(ctx, user) error

**Preconditions**:
- `ctx` must not be nil
- `user` must not be nil
- `user.ID` must be set (UUID v4)
- `user.SiteID` must be set (UUID v4)
- `user.Email` must be valid format and <= 255 chars
- `user.PasswordHash` must be >= 60 chars (bcrypt)
- `user.CreatedAt` and `user.UpdatedAt` must be set

**Postconditions (Success)**:
- User is persisted in repository
- User can be retrieved by `FindByEmail`
- `user.CreatedAt` and `user.UpdatedAt` are preserved
- Returns `nil`

**Error Conditions**:
- `ErrUserAlreadyExists`: User with same email exists in site
- `ErrInvalidUser`: User fails domain validation
- `ErrRepositoryFailure`: Infrastructure error (DB connection, etc.)

**Context Behavior**:
- Respects context cancellation
- Respects context timeout
- Returns `context.Canceled` or `context.DeadlineExceeded` appropriately

### FindByEmail(ctx, siteID, email) (*User, error)

**Preconditions**:
- `ctx` must not be nil
- `siteID` must be valid UUID v4
- `email` must be non-empty string

**Postconditions (Found)**:
- Returns pointer to User entity
- User belongs to specified `siteID`
- User email matches exactly (case-sensitive)
- User passes domain validation
- Returns `nil` error

**Postconditions (Not Found)**:
- Returns `nil, nil` (not found is not an error)
- No partial matches returned
- Site isolation respected

**Error Conditions**:
- `ErrRepositoryFailure`: Infrastructure error only
- Never returns domain validation errors (assumes stored data is valid)

**Context Behavior**:
- Respects context cancellation
- Respects context timeout  
- Returns appropriate context errors

## Error Types

```go
// Domain errors
var (
    ErrUserAlreadyExists = errors.New("user already exists with this email in site")
    ErrInvalidUser = errors.New("user data fails validation")
    ErrRepositoryFailure = errors.New("repository operation failed")
)

// Error wrapping for context
type RepositoryError struct {
    Op   string // Operation name
    Err  error  // Underlying error
    User *User  // Related user (if applicable)
}

func (e *RepositoryError) Error() string {
    return fmt.Sprintf("repository %s: %v", e.Op, e.Err)
}

func (e *RepositoryError) Unwrap() error {
    return e.Err
}
```

## Usage Examples

```go
// Create user
user, err := NewUser(siteID, "user@example.com", hashFromAuth)
if err != nil {
    return fmt.Errorf("invalid user: %w", err)
}

err = repo.Create(ctx, user)
if errors.Is(err, ErrUserAlreadyExists) {
    return fmt.Errorf("email already registered")
}
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Find user  
user, err := repo.FindByEmail(ctx, siteID, "user@example.com")
if err != nil {
    return fmt.Errorf("failed to find user: %w", err)
}
if user == nil {
    return fmt.Errorf("user not found")
}
```

## Implementation Requirements

### Multi-Tenancy
- All operations MUST filter by `site_id`
- Cross-site data access MUST be prevented  
- Queries MUST use `(site_id, email)` for lookups

### Performance
- Create operations SHOULD complete in <100ms p95
- FindByEmail operations SHOULD complete in <50ms p95
- MUST support 1000+ concurrent operations per site

### Security
- Email lookups MUST be case-sensitive exact match
- No user enumeration across sites allowed
- Input validation at domain layer, not repository layer

### Testing
- MUST provide mock implementation for unit testing
- Integration tests MUST use real database
- Contract tests MUST verify all preconditions/postconditions

## Mock Implementation

```go
type MockUserRepository struct {
    users map[string]*User // key: siteID+email
    calls map[string]int   // method call counts
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        users: make(map[string]*User),
        calls: make(map[string]int),
    }
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    m.calls["Create"]++
    
    if ctx.Err() != nil {
        return ctx.Err()
    }
    
    key := user.SiteID.String() + ":" + user.Email
    if _, exists := m.users[key]; exists {
        return ErrUserAlreadyExists
    }
    
    m.users[key] = user
    return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error) {
    m.calls["FindByEmail"]++
    
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    
    key := siteID.String() + ":" + email
    user, exists := m.users[key]
    if !exists {
        return nil, nil
    }
    
    return user, nil
}

// Test helpers
func (m *MockUserRepository) CallCount(method string) int {
    return m.calls[method]
}

func (m *MockUserRepository) Reset() {
    m.users = make(map[string]*User)
    m.calls = make(map[string]int)
}
```

This contract provides a complete specification for implementing the UserRepository interface while maintaining constitutional compliance and clean architecture principles.