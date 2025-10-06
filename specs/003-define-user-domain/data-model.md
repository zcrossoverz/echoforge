# Data Model: User Domain Entity

**Feature**: User Domain Entity and Repository  
**Date**: 2025-10-02  
**Phase**: 1 - Design

## Entity: User

### Core Attributes
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `ID` | UUID | Required, Unique, Immutable | Primary identifier, generated on creation |
| `SiteID` | UUID | Required, Immutable | Site isolation identifier for multi-tenancy |
| `Email` | string | Required, Unique within site, Max 255 chars | User's email address, validated format |
| `PasswordHash` | string | Required, Min 60 chars | bcrypt hash, validated length |
| `CreatedAt` | time.Time | Auto-generated, Immutable | Entity creation timestamp |
| `UpdatedAt` | time.Time | Auto-maintained | Last modification timestamp |

### Validation Rules

#### Email Validation
```go
// Email format validation
const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

// Validation rules:
// - Must match RFC 5322 email format
// - Maximum length: 255 characters
// - Required field, cannot be empty
// - Must be unique within site scope
```

#### Password Hash Validation  
```go
// Password hash validation
const minHashLength = 60 // bcrypt standard

// Validation rules:
// - Minimum length: 60 characters (bcrypt hash size)
// - Required field, cannot be empty
// - Assumed to be pre-hashed with bcrypt
// - Domain layer validates format, not plaintext
```

#### Site ID Validation
```go
// Site ID validation
// - Must be valid UUID v4 format
// - Required field, cannot be empty
// - Immutable after creation
// - Used for tenant isolation
```

### Entity Methods

```go
type User struct {
    ID           uuid.UUID `json:"id"`
    SiteID       uuid.UUID `json:"site_id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// NewUser creates a new User entity with validation
func NewUser(siteID uuid.UUID, email, passwordHash string) (*User, error)

// Validate performs business rule validation
func (u *User) Validate() error

// IsValid returns true if entity passes all validation rules
func (u *User) IsValid() bool
```

### Business Rules

1. **Uniqueness**: Email must be unique within the same site (site_id + email combination)
2. **Immutability**: ID, SiteID, and CreatedAt cannot be changed after creation
3. **Email Format**: Must conform to standard email format regex
4. **Password Security**: Hash must be minimum 60 characters (bcrypt standard)
5. **Required Fields**: All fields except UpdatedAt are required
6. **Tenant Isolation**: All operations must include SiteID for multi-tenant security

### State Transitions

```
[New] → [Valid] → [Persisted]
  |         |
  ↓         ↓
[Invalid] [Error]
```

- **New**: Entity created but not validated
- **Valid**: Entity passes all business rules
- **Invalid**: Entity fails validation rules
- **Persisted**: Entity successfully stored in repository
- **Error**: Repository operation failed

### Repository Interface

```go
type UserRepository interface {
    // Create persists a new user with validation
    Create(ctx context.Context, user *User) error
    
    // FindByEmail retrieves user by email within site
    FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error)
}
```

### Error Types

```go
// Domain errors for business rule violations
type ValidationError struct {
    Field   string
    Message string
}

// Repository errors for persistence failures  
type RepositoryError struct {
    Operation string
    Cause     error
}
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT users_email_length CHECK (char_length(email) <= 255),
    CONSTRAINT users_password_hash_length CHECK (char_length(password_hash) >= 60),
    CONSTRAINT users_site_email_unique UNIQUE (site_id, email)
);

-- Indexes for performance
CREATE INDEX idx_users_site_id ON users(site_id);
CREATE INDEX idx_users_site_email ON users(site_id, email);
```

### Migration Files

**002_create_users_table.up.sql**:
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

**002_create_users_table.down.sql**:
```sql
DROP TABLE IF EXISTS users;
```

## Relationships

### Current (MVP)
- No relationships defined for lean MVP approach
- User is a standalone entity

### Future Extensions
- User → Profile (one-to-one)
- User → Sessions (one-to-many)  
- User → Permissions (many-to-many through roles)
- User → AuditLog (one-to-many)

## Integration Points

### Domain Layer
- `internal/domain/user.go`: Pure entity with validation
- No external dependencies except standard library + UUID

### Use Case Layer  
- `internal/usecase/user_usecase.go`: Business logic using repository interface
- Depends on domain User entity and repository interface

### Adapter Layer
- `adapters/persistence/user_repository.go`: GORM implementation
- Maps domain User to/from database representation
- Handles GORM-specific concerns (transactions, connections)

### Testing
- Domain: Unit tests for entity validation and business rules
- Use Case: Unit tests with mock repository  
- Repository: Integration tests with real database
- Contract: Interface compliance tests

This data model provides the foundation for multi-tenant user management while maintaining clean architecture boundaries and constitutional compliance.