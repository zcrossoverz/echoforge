# Usecase Contracts: Register and Login Authentication

**Date**: October 2, 2025  
**Type**: Go Usecase Interface Contracts  
**Context**: Business logic layer contracts, not HTTP API contracts

## RegisterUsecase Contract

### Interface Definition
```go
type RegisterUsecase interface {
    Execute(ctx context.Context, input RegisterInput) (*domain.User, error)
}
```

### Input Contract
```go
// RegisterInput represents the required data for user registration
type RegisterInput struct {
    SiteID   uuid.UUID `validate:"required,uuid" json:"site_id"`
    Email    string    `validate:"required,email,max=255" json:"email"`  
    Password string    `validate:"required,min=8" json:"password"`
}
```

**Validation Rules**:
- `SiteID`: Required, must be valid UUID
- `Email`: Required, valid email format, maximum 255 characters
- `Password`: Required, minimum 8 characters

### Output Contract
**Success Response**: `*domain.User`
- Returns fully populated User entity with generated ID
- PasswordHash excluded from JSON serialization
- CreatedAt and UpdatedAt timestamps populated

**Error Responses**:
- `ValidationError`: Input validation failures with field-specific messages
- `ErrEmailAlreadyExists`: Email already registered for the specified site
- `context.Canceled`: Context cancellation during execution
- `context.DeadlineExceeded`: Context timeout during execution
- Repository errors: Database connectivity or constraint violations

### Behavior Contract
1. **Input Validation**: Validate all input fields using struct tags
2. **Duplicate Check**: Verify email uniqueness within site using repository
3. **Password Hashing**: Hash password using bcrypt with cost ≥12
4. **User Creation**: Create domain.User entity with generated UUID
5. **Persistence**: Save user via repository with proper error handling
6. **Multi-Tenant**: All operations scoped to provided SiteID

## LoginUsecase Contract

### Interface Definition
```go
type LoginUsecase interface {
    Execute(ctx context.Context, input LoginInput) (*AuthenticationResult, error)
}
```

### Input Contract
```go
// LoginInput represents the required data for user authentication
type LoginInput struct {
    SiteID   uuid.UUID `validate:"required,uuid" json:"site_id"`
    Email    string    `validate:"required,email" json:"email"`
    Password string    `validate:"required" json:"password"`
}
```

**Validation Rules**:
- `SiteID`: Required, must be valid UUID
- `Email`: Required, valid email format
- `Password`: Required (no minimum length for existing users)

### Output Contract
**Success Response**: `*AuthenticationResult`
```go
type AuthenticationResult struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    User      *domain.User `json:"user,omitempty"`
}
```

**Fields**:
- `Token`: JWT string with HS256 signature
- `ExpiresAt`: Token expiration timestamp (24 hours from issuance)
- `User`: Authenticated user entity (optional)

**Error Responses**:
- `ValidationError`: Input validation failures
- `ErrAuthenticationFailed`: Generic authentication failure (security)
- `context.Canceled`: Context cancellation during execution
- `context.DeadlineExceeded`: Context timeout during execution
- Repository errors: Database connectivity issues

### Behavior Contract
1. **Input Validation**: Validate all input fields using struct tags
2. **User Lookup**: Find user by email within specified site
3. **Password Verification**: Compare provided password with stored hash using bcrypt
4. **JWT Generation**: Create signed JWT token with user and site claims
5. **Token Expiration**: Set 24-hour expiration from current time
6. **Security**: Generic error messages prevent user enumeration
7. **Multi-Tenant**: All operations scoped to provided SiteID

## JWT Token Contract

### Claims Structure
```go
type JWTClaims struct {
    UserID string `json:"sub"`      // Subject: User.ID
    SiteID string `json:"site_id"`  // Custom: User.SiteID  
    jwt.RegisteredClaims
}
```

### Standard Claims
- `sub` (subject): User ID as UUID string
- `exp` (expiration): Unix timestamp, 24 hours from issuance
- `iat` (issued at): Unix timestamp of token creation

### Token Format
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Secret**: Configurable via JWT_SECRET environment variable
- **Expiration**: 24 hours from issuance
- **Encoding**: Base64 URL-encoded JWT standard format

## Error Handling Contract

### Validation Errors
```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   any    `json:"value,omitempty"`
}
```

### Authentication Errors
- **Generic Response**: "authentication failed" (no details for security)
- **No User Enumeration**: Same error for "user not found" vs "wrong password"
- **Rate Limiting**: Deferred to HTTP middleware layer

### Repository Errors
- **Connection Failures**: Propagated as infrastructure errors
- **Constraint Violations**: Converted to domain-specific errors
- **Context Handling**: Timeout and cancellation support

## Testing Contracts

### Unit Test Requirements
- **Mock Repository**: All database interactions mocked
- **Input Validation**: Test all validation rules and edge cases
- **Error Scenarios**: Test all error conditions and edge cases
- **Context Handling**: Test cancellation and timeout scenarios
- **Security**: Verify password hashing and JWT generation

### Test Coverage Requirements
- **Minimum Coverage**: 80% per constitutional requirements
- **Branch Coverage**: All conditional logic paths tested
- **Error Paths**: All error conditions covered
- **Edge Cases**: Boundary conditions and invalid inputs

### Integration Test Requirements
- **Real Repository**: Test with actual database connections
- **End-to-End**: Full usecase execution with real dependencies
- **Multi-Tenant**: Verify site isolation in integration scenarios
- **Performance**: Validate response time requirements (<2 seconds)

---

**Contract Status**: COMPLETE  
**Interfaces Defined**: 2 usecase interfaces  
**Validation Rules**: Comprehensive input/output contracts  
**Ready for Contract Test Generation**: YES