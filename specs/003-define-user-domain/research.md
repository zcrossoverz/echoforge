# Research Report: User Domain Entity and Repository

**Feature**: User Domain Entity and Repository  
**Date**: 2025-10-02  
**Research Phase**: Phase 0

## Technical Decisions

### 1. Domain Entity Design Pattern
**Decision**: Pure domain entity with value object validation  
**Rationale**: 
- Hexagonal architecture requires domain layer to be framework-agnostic
- User entity contains business rules (email validation, password hash requirements)
- Validation belongs in domain layer, not infrastructure
- Supports testability and maintainability

**Alternatives considered**:
- GORM struct with tags: Rejected - couples domain to infrastructure
- Separate validation service: Rejected - adds unnecessary complexity for MVP

### 2. Repository Pattern Implementation  
**Decision**: Interface in domain, GORM implementation in adapters
**Rationale**:
- Hexagonal architecture requires ports (interfaces) in domain
- GORM implementation belongs in persistence adapter
- Enables testing with mock repositories
- Supports future database migrations or alternative ORMs

**Alternatives considered**:
- Direct GORM usage in use cases: Rejected - violates hexagonal architecture
- Generic repository: Rejected - YAGNI for MVP, adds complexity

### 3. Multi-Tenant Site Isolation Strategy
**Decision**: site_id field in User entity and all repository queries
**Rationale**:
- Constitutional requirement for tenant isolation
- Prevents accidental cross-site data access
- Enables shared database with logical separation
- Supports horizontal scaling by sharding on site_id

**Alternatives considered**:
- Separate databases per site: Rejected - increases operational complexity
- Context-based filtering: Rejected - too easy to bypass accidentally

### 4. Password Storage Security  
**Decision**: Require pre-hashed passwords with bcrypt validation
**Rationale**:
- Domain entity should not handle plaintext passwords
- Password hashing belongs in authentication layer
- Domain validates hash format (60+ characters for bcrypt)
- Follows security best practices

**Alternatives considered**:
- Hash in domain entity: Rejected - crypto belongs in security layer  
- Store plaintext: Rejected - security violation

### 5. Error Handling Strategy
**Decision**: Domain errors for validation, repository errors for persistence
**Rationale**:
- Clear separation between business rule violations and infrastructure failures
- Enables different error handling strategies (validation vs retry logic)
- Supports detailed error messages for user feedback

**Alternatives considered**:
- Single error type: Rejected - loses context for error handling
- Error codes: Rejected - Go idioms prefer typed errors

### 6. UUID vs Auto-increment IDs
**Decision**: UUID v4 for user IDs
**Rationale**:
- Prevents ID enumeration attacks across sites
- Enables distributed system scaling
- No collision risk with site_id isolation
- Constitutional security requirement

**Alternatives considered**:
- Auto-increment: Rejected - security risk, not scalable
- UUID v1: Rejected - contains MAC address information

## Implementation Patterns

### Domain Layer Patterns
- **Value Objects**: Email validation with format checking
- **Entity Validation**: Self-validating User entity
- **Domain Events**: Future extensibility for user creation events

### Repository Patterns  
- **Interface Segregation**: Minimal repository interface (Create, FindByEmail)
- **Context Usage**: All operations accept context.Context for cancellation
- **Error Wrapping**: Repository errors wrapped with context

### Testing Patterns
- **Table-Driven Tests**: For validation scenarios
- **Mock Repository**: For use case testing
- **Integration Tests**: Real database testing for repository

## Dependencies Analysis

### Core Dependencies (already in go.mod)
- `github.com/google/uuid`: UUID generation - ✅ Available v1.6.0
- `gorm.io/gorm`: ORM for persistence - ✅ Available v1.25.12
- `github.com/stretchr/testify`: Testing framework - ✅ Available v1.11.1
- `golang.org/x/crypto/bcrypt`: Password hashing validation - ✅ Available v0.42.0

### Validation Dependencies
- `github.com/go-playground/validator/v10`: Struct validation - ✅ Available v10.22.1

No additional dependencies required for MVP implementation.

## Performance Considerations

### Database Queries
- Index on (site_id, email) for fast user lookup
- site_id filter on all queries prevents full table scans
- Prepared statements through GORM for query caching

### Memory Usage
- Minimal User struct size (ID, SiteID, Email, PasswordHash, timestamps)
- No eager loading of relationships (none defined yet)
- Context cancellation prevents resource leaks

### Concurrency
- Repository methods are goroutine-safe through GORM
- No shared mutable state in domain entities
- Context-based timeout handling

## Security Analysis

### Input Validation
- Email format validation using regex
- Password hash length validation (bcrypt = 60 chars)
- Site ID UUID format validation

### Data Protection
- No plaintext password storage
- Site isolation prevents cross-tenant access
- UUIDs prevent ID enumeration

### OWASP Top 10 Compliance
- A01 Broken Access Control: Prevented by site_id isolation
- A02 Cryptographic Failures: bcrypt hash validation
- A03 Injection: Prevented by GORM parameterized queries
- A05 Security Misconfiguration: Input validation prevents bad data

## Migration Strategy

### Database Schema
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    site_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(site_id, email)
);

CREATE INDEX idx_users_site_email ON users(site_id, email);
```

### Backward Compatibility
- New table, no existing data to migrate
- Interface-based design allows implementation changes
- Constitutional SemVer compliance maintained

## Outstanding Questions
*All clarifications resolved based on feature specification*

- ✅ Email uniqueness scope: Within site only (confirmed in FR-007)
- ✅ Password storage format: Pre-hashed with bcrypt (confirmed in security requirements)  
- ✅ User lifecycle: Create and retrieve only for MVP (confirmed in scope)
- ✅ Validation rules: Email format + password hash length (confirmed in FR-003, FR-004)

## Next Steps for Phase 1
1. Design User entity struct with validation methods
2. Define Repository interface with Create and FindByEmail methods  
3. Create data model documentation
4. Generate database migration files
5. Define contract tests for repository interface
6. Create quickstart guide for user entity usage