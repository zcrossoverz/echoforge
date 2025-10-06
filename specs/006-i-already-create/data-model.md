# Data Model: Database Connection and Authentication APIs

**Feature**: Database Connection and Authentication APIs  
**Date**: October 4, 2025  
**Status**: Complete

## Entity Definitions

### 1. User Entity

**Purpose**: Represents a registered user account in the bloggo system

**Fields**:
- `ID` (UUID): Primary key, auto-generated
- `Email` (string): Unique identifier, max 320 characters, RFC compliant
- `PasswordHash` (string): bcrypt hashed password, never store plaintext
- `CreatedAt` (timestamp): Account creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Email: required, valid email format, unique within database
- Password: minimum 8 characters, at least one letter and number (pre-hash)
- All fields: non-empty, trimmed whitespace

**Business Rules**:
- Email uniqueness enforced at database level
- Password must be hashed with bcrypt cost factor 12
- Created/Updated timestamps managed automatically
- Soft delete not implemented (hard delete for GDPR compliance)

**GORM Model**:
```go
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Email        string    `gorm:"type:varchar(320);uniqueIndex;not null" json:"email" validate:"required,email,max=320"`
    PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

### 2. Authentication Session (Stateless)

**Purpose**: Represents user session state via JWT tokens (no database storage)

**JWT Payload Fields**:
- `user_id` (UUID): Reference to User.ID
- `email` (string): User email for convenience
- `iat` (timestamp): Token issued at time
- `exp` (timestamp): Token expiration time
- `iss` (string): Token issuer ("bloggo")

**Validation Rules**:
- Token signature must be valid
- Token must not be expired
- User ID must exist in database
- Token must not be blacklisted (logout functionality)

**Business Rules**:
- Tokens expire after 24 hours
- No refresh token for MVP (re-login required)
- Logout adds token to blacklist until expiration
- One active session per user (new login invalidates old tokens)

## Database Schema

### Tables

#### users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### auth_blacklist (for logout functionality)
```sql
CREATE TABLE auth_blacklist (
    token_id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Indexes

```sql
-- Primary indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Blacklist indexes
CREATE INDEX idx_auth_blacklist_user_id ON auth_blacklist(user_id);
CREATE INDEX idx_auth_blacklist_expires_at ON auth_blacklist(expires_at);
```

### Migrations

Migration files will be created in `migrations/` directory:
- `001_create_users_table.up.sql` - Create users table
- `001_create_users_table.down.sql` - Drop users table
- `002_create_auth_blacklist_table.up.sql` - Create blacklist table
- `002_create_auth_blacklist_table.down.sql` - Drop blacklist table

## Data Access Patterns

### User Repository Interface

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Authentication Repository Interface

```go
type AuthRepository interface {
    BlacklistToken(ctx context.Context, tokenID string, userID uuid.UUID, expiresAt time.Time) error
    IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
    CleanupExpiredTokens(ctx context.Context) error
}
```

## State Transitions

### User Lifecycle
1. **Registration**: Anonymous → Registered User
2. **Login**: Registered User → Authenticated User
3. **Logout**: Authenticated User → Registered User
4. **Account Deletion**: Registered User → Deleted (hard delete)

### Authentication Flow
1. **Register**: Create user account, return auth token
2. **Login**: Validate credentials, generate new token
3. **Logout**: Add current token to blacklist
4. **Token Refresh**: Not implemented in MVP (future feature)

## Validation and Constraints

### Database Constraints
- Email uniqueness enforced at database level
- Foreign key constraints for referential integrity
- Not null constraints on required fields
- Check constraints for data validation where possible

### Application Constraints
- Password complexity validated before hashing
- Email format validation using go-playground/validator
- Rate limiting enforced at application level
- Input sanitization to prevent XSS/injection

## Data Relationships

### Current MVP
- No complex relationships (single entity system)
- User entity is standalone
- Authentication state managed via JWT (stateless)

### Future Extensions (Post-MVP)
- User → Blog Posts (one-to-many)
- User → Comments (one-to-many)  
- User → Categories (many-to-many)
- User → User Roles (many-to-many)

## Performance Considerations

### Database Optimization
- Primary key (UUID) for fast lookups
- Email index for authentication queries
- Connection pooling for concurrent access
- Query optimization with proper where clauses

### Memory Management
- Minimal in-memory state (stateless JWT)
- Periodic blacklist cleanup to prevent memory leaks
- Connection pool limits to prevent resource exhaustion

### Caching Strategy (Future)
- User profile caching (Redis)
- Blacklist token caching for faster lookup
- Database query result caching for read-heavy operations

## Security Considerations

### Data Protection
- Passwords never stored in plaintext
- JWT secrets stored in environment variables
- Database credentials in configuration files (not code)
- Input validation prevents SQL injection

### Access Control
- Authentication required for protected endpoints
- Authorization checks before data access
- Rate limiting prevents brute force attacks
- Audit logging for security events

### Compliance
- GDPR: Hard delete capability for user accounts
- OWASP: Input validation, secure authentication
- Constitutional: bcrypt + JWT requirements met

**Data Model Status**: COMPLETE - Ready for contract generation