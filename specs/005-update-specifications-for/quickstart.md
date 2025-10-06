# Quickstart: Clone-and-Extend User Domain

**Date**: October 4, 2025 | **Feature**: Update User Domain and Authentication

## Overview
This guide walks through implementing the updated user domain and authentication system for the clone-and-extend architectural model, removing multi-tenant `site_id` isolation in favor of separate database instances per site.

## Prerequisites
- Go 1.25+
- PostgreSQL 16+
- Existing echoforge codebase with hexagonal architecture
- Understanding of TDD methodology

## Implementation Steps

### 1. Update Domain Entity
**File**: `internal/domain/user.go`

```go
// Remove SiteID field from User struct
type User struct {
    ID           uuid.UUID `json:"id"`
    // SiteID    uuid.UUID `json:"site_id"` // REMOVE THIS LINE
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Update NewUser constructor
func NewUser(email, passwordHash string) (*User, error) {
    // Remove siteID parameter
    user := &User{
        ID:           uuid.New(),
        // SiteID:    siteID, // REMOVE THIS LINE
        Email:        email,
        PasswordHash: passwordHash,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    // Validation logic unchanged
    return user, user.Validate()
}

// Update Validate method
func (u *User) Validate() error {
    // Remove SiteID validation
    // if u.SiteID == uuid.Nil { ... } // REMOVE THIS BLOCK
    
    // Keep all other validation logic unchanged
    return nil
}
```

### 2. Update Repository Interface
**File**: `internal/domain/user.go`

```go
// Update UserRepository interface
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    
    // Remove siteID parameter from FindByEmail
    FindByEmail(ctx context.Context, email string) (*User, error)
}
```

### 3. Update Use Case Input DTOs
**File**: `internal/usecase/user/register.go`

```go
// Update RegisterInput struct
type RegisterInput struct {
    // SiteID uuid.UUID `json:"site_id" validate:"required"` // REMOVE THIS LINE
    Email    string `json:"email" validate:"required,email,max=320"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}

// Update Execute method
func (uc *RegisterUsecaseImpl) Execute(ctx context.Context, input RegisterInput) (*domain.User, error) {
    // Remove siteID validation and usage
    // Validation logic simplified - no siteID parameter needed
    
    // Create user without siteID
    user, err := domain.NewUser(input.Email, hashedPassword)
    if err != nil {
        return nil, err
    }
    
    // Check existing user without siteID scope
    existingUser, err := uc.userRepo.FindByEmail(ctx, input.Email)
    // Rest of logic unchanged
}
```

**File**: `internal/usecase/user/login.go`

```go
// Update LoginInput struct
type LoginInput struct {
    // SiteID uuid.UUID `json:"site_id" validate:"required"` // REMOVE THIS LINE
    Email    string `json:"email" validate:"required,email,max=320"`
    Password string `json:"password" validate:"required,max=128"`
}

// Update Execute method
func (uc *LoginUsecaseImpl) Execute(ctx context.Context, input LoginInput) (*AuthenticationResult, error) {
    // Find user without siteID scope
    user, err := uc.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, err
    }
    
    // Generate token without siteID
    token, expiresAt, err := auth.GenerateToken(user.ID, uc.jwtSecret)
    // Rest of logic unchanged
}
```

### 4. Update JWT Authentication
**File**: `pkg/auth/jwt.go`

```go
// Update JWTClaims struct
type JWTClaims struct {
    UserID string `json:"sub"` // Keep user ID
    // SiteID string `json:"site_id"` // REMOVE THIS LINE
    jwt.RegisteredClaims
}

// Update GenerateToken function
func GenerateToken(userID uuid.UUID, secret string) (string, time.Time, error) {
    // Remove siteID parameter
    expirationTime := time.Now().Add(24 * time.Hour)
    
    claims := &JWTClaims{
        UserID: userID.String(),
        // SiteID: siteID.String(), // REMOVE THIS LINE
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    // Token generation logic unchanged
}
```

### 5. Update Tests
Apply TDD approach - update tests first, then verify implementation:

**Domain Tests**: Remove `siteID` parameters from all test cases
**Use Case Tests**: Remove site isolation test scenarios  
**JWT Tests**: Verify tokens contain only user ID claims

### 6. Database Migration (Optional)
For new deployments, update schema:

```sql
-- Option A: Keep site_id column for compatibility (recommended)
ALTER TABLE users ALTER COLUMN site_id DROP NOT NULL;
DROP INDEX IF EXISTS idx_users_site_email;
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- Option B: Remove site_id column (breaking change)
-- ALTER TABLE users DROP COLUMN site_id;
-- CREATE UNIQUE INDEX idx_users_email ON users(email);
```

## Testing Strategy

### 1. Unit Tests
```bash
# Run domain tests
go test ./internal/domain/... -v

# Run use case tests  
go test ./internal/usecase/user/... -v

# Run auth tests
go test ./pkg/auth/... -v
```

### 2. Integration Tests
```bash
# Run all tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 3. Validation Checklist
- [ ] User entity validates without `SiteID`
- [ ] Registration works without site context
- [ ] Login works without site context  
- [ ] JWT tokens contain only user ID
- [ ] Repository methods work without `siteID` parameters
- [ ] All tests pass with 80%+ coverage

## Configuration Changes

### Environment Variables
```bash
# No changes needed - JWT_SECRET unchanged
JWT_SECRET=your-secret-key

# Database per site clone (each site has separate config)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=echoforge_site1  # Unique per site clone
DB_USER=postgres
DB_PASSWORD=password
```

### Clone Deployment
Each site clone needs:
1. Separate Git repository (cloned from core)
2. Separate database instance
3. Separate configuration files
4. Independent deployment pipeline

## Performance Benefits
- **Query Performance**: Eliminated `site_id` JOIN conditions
- **Index Efficiency**: Simpler unique constraints
- **Token Size**: Smaller JWT tokens (no site_id claim)
- **Memory Usage**: Reduced struct sizes and validation overhead

## Migration Path
1. **New Sites**: Use updated architecture from start
2. **Existing Sites**: 
   - Create separate database per site
   - Migrate data with site-specific queries
   - Deploy updated code to each site clone
   - Update configuration and environment variables

## Troubleshooting

### Common Issues
1. **Compilation Errors**: Ensure all `siteID` parameters removed
2. **Test Failures**: Update test expectations to remove site context
3. **Database Constraints**: Verify unique email constraint applied
4. **JWT Validation**: Ensure token validation doesn't expect `site_id` claim

### Validation Commands
```bash
# Check for remaining site_id references
grep -r "site_id\|siteID\|SiteID" internal/ pkg/ --exclude-dir=.git

# Verify test coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "total:"
```

## Next Steps
1. Follow TDD approach: Write tests first
2. Implement changes incrementally
3. Validate each component independently
4. Run integration tests continuously
5. Update deployment documentation
6. Plan data migration for existing multi-tenant deployments