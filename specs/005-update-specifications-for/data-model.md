# Data Model: Clone-and-Extend User Domain

**Date**: October 4, 2025 | **Feature**: Update User Domain and Authentication

## Entity Definitions

### User Entity (Updated)
```go
// User represents a registered user (site_id removed for clone-and-extend model)
type User struct {
    ID           uuid.UUID `json:"id"`
    // SiteID    uuid.UUID `json:"site_id"` // REMOVED: Each clone has separate DB
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**Changes from Current**:
- ❌ **Removed**: `SiteID uuid.UUID` field
- ✅ **Kept**: All other fields unchanged for backward compatibility
- ✅ **Validation**: Email uniqueness enforced at database level (no site scope needed)

### Input DTOs (Updated)

#### RegisterInput
```go
type RegisterInput struct {
    // SiteID uuid.UUID `json:"site_id" validate:"required"` // REMOVED
    Email    string `json:"email" validate:"required,email,max=320"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}
```

#### LoginInput
```go
type LoginInput struct {
    // SiteID uuid.UUID `json:"site_id" validate:"required"` // REMOVED
    Email    string `json:"email" validate:"required,email,max=320"`
    Password string `json:"password" validate:"required,max=128"`
}
```

**Changes from Current**:
- ❌ **Removed**: `SiteID` fields and validation tags
- ✅ **Kept**: Email and password validation unchanged

### JWT Claims (Updated)
```go
type JWTClaims struct {
    UserID string `json:"sub"`     // Subject: User ID
    // SiteID string `json:"site_id"` // REMOVED: No site context needed
    jwt.RegisteredClaims
}
```

**Changes from Current**:
- ❌ **Removed**: `SiteID string` claim
- ✅ **Kept**: `UserID string` as subject claim
- ✅ **Kept**: Standard registered claims (exp, iat, etc.)

## Repository Interfaces (Updated)

### UserRepository
```go
type UserRepository interface {
    // Create persists a new user entity
    Create(ctx context.Context, user *User) error

    // FindByEmail retrieves a user by email (site_id parameter removed)
    FindByEmail(ctx context.Context, email string) (*User, error)
}
```

**Changes from Current**:
- ❌ **Removed**: `siteID uuid.UUID` parameter from `FindByEmail()`
- ✅ **Kept**: Context-first parameter pattern
- ✅ **Kept**: Error handling approach

## Database Schema Impact

### User Table (Existing)
```sql
-- Current schema (no changes needed for clone-and-extend)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- site_id UUID NOT NULL,  -- Column can remain but unused in clone model
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints updated for clone-and-extend
    UNIQUE(email)  -- Was: UNIQUE(site_id, email)
);
```

**Migration Strategy**:
- **Option A**: Keep `site_id` column for backward compatibility (unused in clone model)
- **Option B**: Drop `site_id` column in new migrations (breaking change)
- **Recommendation**: Option A for this refactoring (can be dropped in future version)

## Validation Rules

### Domain Validation (Updated)
```go
func (u *User) Validate() error {
    // Required fields
    if u.ID == uuid.Nil {
        return fmt.Errorf("ID: %w", ErrRequiredField)
    }
    // if u.SiteID == uuid.Nil {  // REMOVED
    //     return fmt.Errorf("SiteID: %w", ErrRequiredField)
    // }
    if u.Email == "" {
        return fmt.Errorf("Email: %w", ErrRequiredField)
    }
    if u.PasswordHash == "" {
        return fmt.Errorf("PasswordHash: %w", ErrRequiredField)
    }

    // Email validation (unchanged)
    if len(u.Email) > MaxEmailLength {
        return fmt.Errorf("Email: %w (%d > %d)", ErrEmailTooLong, len(u.Email), MaxEmailLength)
    }
    if !emailRegex.MatchString(u.Email) {
        return fmt.Errorf("Email: %w", ErrInvalidEmail)
    }

    // Password hash validation (unchanged)
    if len(u.PasswordHash) < MinPasswordLength {
        return fmt.Errorf("PasswordHash: %w (%d < %d)", ErrPasswordHashTooShort, len(u.PasswordHash), MinPasswordLength)
    }

    return nil
}
```

### Input Validation (go-playground/validator)
- **RegisterInput**: `email` and `password` validation (site_id removed)
- **LoginInput**: `email` and `password` validation (site_id removed)

## Business Rules Impact

### User Uniqueness
- **Before**: Email unique per site (`UNIQUE(site_id, email)`)
- **After**: Email globally unique (`UNIQUE(email)`)
- **Rationale**: Each clone operates independent user base, global uniqueness simpler

### Authentication Scope
- **Before**: User authenticated within site context
- **After**: User authenticated within clone instance
- **Rationale**: Clone-and-extend model provides natural isolation

### Repository Operations
- **Before**: All queries include `site_id` filter
- **After**: All queries operate on entire clone database
- **Rationale**: Database-level isolation eliminates need for application-level filtering

## Performance Implications

### Query Performance
- **Improved**: Removal of `site_id` JOIN conditions
- **Improved**: Simpler indexes without compound `(site_id, email)` keys
- **Improved**: Reduced parameter validation overhead

### Memory Usage
- **Reduced**: Smaller JWT tokens (no site_id claim)
- **Reduced**: Fewer struct fields in DTOs
- **Reduced**: Simplified validation logic

### Database Connections
- **Unchanged**: Each clone maintains its own connection pool
- **Benefit**: Natural isolation without connection multiplexing complexity

## Compatibility Matrix

| Component | Current (Multi-tenant) | Updated (Clone-and-Extend) | Breaking Change |
|-----------|------------------------|---------------------------|-----------------|
| User Entity | Has `SiteID` | No `SiteID` | ❌ Yes |
| RegisterInput | Has `SiteID` | No `SiteID` | ❌ Yes |
| LoginInput | Has `SiteID` | No `SiteID` | ❌ Yes |
| JWT Claims | Has `SiteID` | No `SiteID` | ❌ Yes |
| UserRepository | Site-scoped | Clone-scoped | ❌ Yes |
| Database Schema | Compound unique key | Simple unique key | ⚠️ Optional |

**Migration Path**: This is an architectural refactoring with acceptable breaking changes for long-term benefits.