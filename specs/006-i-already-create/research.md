# Research: Database Connection and Authentication APIs

**Feature**: Database Connection and Authentication APIs  
**Date**: October 4, 2025  
**Status**: Complete

## Research Tasks Completed

### 1. Database Connection with GORM and PostgreSQL

**Decision**: Use GORM v1.26+ with PostgreSQL driver and connection pooling  
**Rationale**: 
- Constitutional requirement for GORM v1.26+
- PostgreSQL 16+ provides excellent performance and reliability
- Connection pooling essential for 1000+ concurrent users
- Auto-migration capabilities for zero-setup deployment

**Implementation Approach**:
- Connection string: `postgres://postgres:admin@localhost:5432/bloggo?sslmode=disable`
- Connection pool: max open connections = 25, max idle = 10
- Auto-migration on startup for user tables
- Health check endpoint to verify DB connectivity

**Alternatives Considered**:
- Direct database/sql: Rejected due to constitutional requirement for GORM
- Other ORMs (like ent): Rejected due to constitutional mandate

### 2. Password Security Requirements

**Decision**: bcrypt with cost factor 12, minimum password length 8 characters  
**Rationale**:
- Constitutional requirement for bcrypt
- Cost factor 12 balances security vs performance (≤200ms per hash)
- 8-character minimum aligns with modern security practices
- No maximum length limit to avoid user frustration

**Implementation Approach**:
- Use golang.org/x/crypto/bcrypt package
- Store only hashed passwords, never plaintext
- Validate password complexity client-side for UX, server-side for security

**Alternatives Considered**:
- Argon2: Better algorithm but constitutional requirement is bcrypt
- Lower cost factor: Insufficient security for production use

### 3. JWT Token Management

**Decision**: HS256 algorithm with 24-hour expiration, refresh token pattern  
**Rationale**:
- HS256 simpler than RS256 for single-service architecture
- 24-hour expiration balances security vs user experience
- Refresh tokens enable secure long-term sessions
- Token revocation via blacklist for logout functionality

**Implementation Approach**:
- Use github.com/golang-jwt/jwt/v5 package
- JWT payload: user_id, email, issued_at, expires_at
- Store JWT secret in environment variable
- Middleware for token validation on protected routes

**Alternatives Considered**:
- RS256 with public/private keys: Overkill for single-service MVP
- Longer expiration: Security risk for stolen tokens
- Stateful sessions: Against constitutional preference for stateless auth

### 4. Rate Limiting Strategy

**Decision**: Token bucket algorithm with 5 attempts per minute per IP  
**Rationale**:
- Prevents brute force attacks while allowing legitimate retries
- IP-based limiting simple to implement and effective
- 5 attempts sufficient for legitimate users with typos
- 1-minute window allows quick retry after rate limit hit

**Implementation Approach**:
- Use golang.org/x/time/rate package for token bucket
- Store rate limit state in memory (Redis for production scaling)
- Return 429 status with Retry-After header
- Apply to login and registration endpoints

**Alternatives Considered**:
- Account-based limiting: Complex to implement, can be gamed
- CAPTCHA: Adds friction, not needed for MVP
- Stricter limits: May frustrate legitimate users

### 5. Database Schema Design

**Decision**: Single users table with minimal required fields  
**Rationale**:
- Lean MVP approach per constitution
- Email as unique identifier per constitutional requirement
- UUID for user ID to prevent enumeration attacks
- Timestamps for audit trail and potential cleanup

**Schema Structure**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Alternatives Considered**:
- Additional user profile fields: Deferred for MVP
- Separate authentication table: Unnecessary complexity for MVP
- Integer IDs: UUIDs prevent enumeration attacks

### 6. API Endpoint Design

**Decision**: RESTful endpoints under /api/v1/ with standard HTTP methods  
**Rationale**:
- Constitutional requirement for versioned APIs
- REST familiar to developers and well-documented
- Standard HTTP status codes for clear communication
- JSON request/response for modern API expectations

**Endpoint Structure**:
- POST /api/v1/auth/register - User registration
- POST /api/v1/auth/login - User authentication
- POST /api/v1/auth/logout - Session termination
- GET /api/v1/auth/profile - Get user profile (protected)
- GET /api/v1/health - System health check

**Alternatives Considered**:
- GraphQL: Overkill for simple auth operations
- RPC-style endpoints: Less standard than REST
- Non-versioned endpoints: Against constitutional requirements

### 7. Input Validation Strategy

**Decision**: go-playground/validator/v10 with custom validation rules  
**Rationale**:
- Constitutional requirement for input validation
- Struct-based validation integrates well with Go
- Comprehensive validation rules available
- Custom validators for business logic (e.g., password strength)

**Validation Rules**:
- Email: RFC compliant, max 320 characters
- Password: minimum 8 characters, at least one letter and number
- All inputs: trimmed of whitespace, XSS prevention

**Alternatives Considered**:
- Manual validation: Error-prone and maintenance burden
- Other validation libraries: validator/v10 is industry standard

### 8. Error Handling and Security Logging

**Decision**: Structured error responses with security event logging  
**Rationale**:
- OWASP compliance requires audit logging
- Structured errors improve API usability
- Security events enable monitoring and alerting
- Avoid information disclosure in error messages

**Implementation Approach**:
- Use existing Zap logger for structured security events
- Generic error messages to external users
- Detailed logging for security team monitoring
- Correlation IDs for request tracing

**Alternatives Considered**:
- Detailed error messages: Security risk (information disclosure)
- No security logging: Against OWASP and constitutional requirements

## Resolution of NEEDS CLARIFICATION Items

From the original specification, all clarification items have been resolved:

1. **Password complexity rules**: Minimum 8 characters, at least one letter and number
2. **Rate limiting rules**: 5 attempts per minute per IP address
3. **User roles/permissions**: None for MVP (deferred to future iterations)
4. **Password reset mechanism**: Deferred to future iteration (not MVP requirement)
5. **Session timeout duration**: 24-hour JWT expiration with refresh capability
6. **Data encryption beyond passwords**: Only passwords encrypted (bcrypt), other data in transit via HTTPS
7. **Error handling preferences**: Generic messages to users, detailed security logging
8. **Integration requirements**: Standalone auth system, no external integrations for MVP

## Dependencies and Versions

All dependencies align with constitutional requirements:

- **GORM**: v1.26.12 (ORM and migrations)
- **Gin**: v1.10.0 (HTTP framework)
- **JWT**: github.com/golang-jwt/jwt/v5 v5.3.0
- **Bcrypt**: golang.org/x/crypto (password hashing)
- **Validator**: github.com/go-playground/validator/v10 v10.27.0
- **Rate Limiting**: golang.org/x/time/rate (built-in)
- **PostgreSQL Driver**: github.com/jackc/pgx/v5 (via GORM)

## Performance Considerations

- Database connection pooling for concurrent access
- JWT stateless authentication reduces server memory
- Bcrypt cost factor 12 provides <200ms hash time
- Rate limiting prevents resource exhaustion
- Indexed email lookups for fast authentication

## Security Measures

- OWASP Top 10 compliance through input validation and secure defaults
- Password hashing with bcrypt (never store plaintext)
- JWT tokens with reasonable expiration
- Rate limiting against brute force attacks
- Security event logging for monitoring
- HTTPS enforcement (configuration level)

**Research Status**: COMPLETE - Ready for Phase 1 Design