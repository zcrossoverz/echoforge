# Data Model: Register and Login Authentication Usecases

**Date**: October 2, 2025  
**Feature**: Register and Login Authentication  
**Dependencies**: Existing User domain entity from Task 1.2

## Entity Overview

### RegisterInput
**Purpose**: Input data transfer object for user registration usecase  
**Location**: `internal/usecase/user/register.go`

```go
type RegisterInput struct {
    SiteID   uuid.UUID `validate:"required,uuid" json:"site_id"`
    Email    string    `validate:"required,email,max=255" json:"email"`
    Password string    `validate:"required,min=8" json:"password"`
}
```

**Fields**:
- `SiteID`: UUID identifying the tenant site (required, must be valid UUID)
- `Email`: User email address (required, valid email format, max 255 chars)
- `Password`: Plain text password (required, minimum 8 characters)

**Validation Rules**:
- All fields required (no nil/empty values)
- Email must be valid RFC 5322 format
- Email length ≤ 255 characters (database constraint)
- Password minimum 8 characters (OWASP recommendation)
- SiteID must be valid UUID format

**State Transitions**: N/A (stateless input DTO)

### LoginInput
**Purpose**: Input data transfer object for user authentication usecase  
**Location**: `internal/usecase/user/login.go`

```go
type LoginInput struct {
    SiteID   uuid.UUID `validate:"required,uuid" json:"site_id"`
    Email    string    `validate:"required,email" json:"email"`
    Password string    `validate:"required" json:"password"`
}
```

**Fields**:
- `SiteID`: UUID identifying the tenant site (required, must be valid UUID)
- `Email`: User email address (required, valid email format)
- `Password`: Plain text password for verification (required)

**Validation Rules**:
- All fields required (no nil/empty values)
- Email must be valid format (no length validation needed for lookup)
- Password required but no minimum length (user may have old password)
- SiteID must be valid UUID format

**State Transitions**: N/A (stateless input DTO)

### AuthenticationResult
**Purpose**: Output data transfer object for successful login  
**Location**: `internal/usecase/user/login.go`

```go
type AuthenticationResult struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    User      *domain.User `json:"user,omitempty"`
}
```

**Fields**:
- `Token`: JWT authentication token string
- `ExpiresAt`: Token expiration timestamp
- `User`: Authenticated user entity (optional, may be omitted for security)

**Validation Rules**:
- Token must be valid JWT string
- ExpiresAt must be future timestamp
- User must be valid domain entity if included

**State Transitions**: N/A (stateless output DTO)

## Existing Entities (Referenced)

### User (from domain layer)
**Purpose**: Core user domain entity with business rules  
**Location**: `internal/domain/user.go` (existing from Task 1.2)  
**Relationship**: Target of registration, source of authentication

```go
type User struct {
    ID           uuid.UUID `json:"id"`
    SiteID       uuid.UUID `json:"site_id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // Excluded from JSON serialization
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**Key Constraints**:
- ID and SiteID must be valid UUIDs
- Email unique within site (enforced by repository/database)
- PasswordHash minimum 60 characters (bcrypt requirement)
- Multi-tenant isolation via SiteID

### UserRepository (interface)
**Purpose**: Data persistence abstraction for user operations  
**Location**: `internal/domain/user.go` (existing interface)  
**Relationship**: Dependency for both usecases

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, siteID uuid.UUID, email string) (*User, error)
}
```

**Operations Used**:
- `Create`: Used by RegisterUsecase to persist new users
- `FindByEmail`: Used by both usecases (duplicate check + authentication)

## JWT Token Claims

### Standard Claims
- `sub` (subject): User.ID as string
- `exp` (expiration): Unix timestamp 24 hours from issuance
- `iat` (issued at): Unix timestamp of token creation

### Custom Claims
- `site_id`: User.SiteID as string for multi-tenant authorization

**Example JWT Payload**:
```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "site_id": "987fcdeb-51a2-43d7-b123-456789abcdef",
  "exp": 1696291200,
  "iat": 1696204800
}
```

## Error Types

### ValidationError
**Purpose**: Structured validation failure information  
**Usage**: Both usecases for input validation failures

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   any    `json:"value,omitempty"`
}
```

### AuthenticationError
**Purpose**: Generic authentication failure (no details for security)  
**Usage**: LoginUsecase for credential failures

```go
var ErrAuthenticationFailed = errors.New("authentication failed")
```

### RegistrationError
**Purpose**: Registration-specific failures  
**Usage**: RegisterUsecase for business rule violations

```go
var ErrEmailAlreadyExists = errors.New("email address already registered")
```

## Data Flow

### Registration Flow
1. `RegisterInput` → Validation → `domain.User` creation
2. Repository duplicate check via `FindByEmail`
3. Password hashing with bcrypt
4. User persistence via `Create`
5. Return created `domain.User` entity

### Login Flow
1. `LoginInput` → Validation → Repository lookup via `FindByEmail`
2. Password verification with bcrypt
3. JWT token generation with claims
4. Return `AuthenticationResult` with token and expiration

## Database Interactions

**Tables Used**: `users` (existing from Task 1.2)  
**Operations**: 
- INSERT (registration)
- SELECT WHERE site_id = ? AND email = ? (both usecases)

**Multi-Tenant Isolation**: All database queries include `site_id` filter  
**Constraints Leveraged**: Unique constraint on (site_id, email) for duplicate prevention

---

**Design Status**: COMPLETE  
**Entities Defined**: 3 new DTOs + 2 existing domain entities  
**Validation Rules**: Specified for all inputs  
**Ready for Contract Generation**: YES