# Research: Register and Login Authentication Usecases

**Date**: October 2, 2025  
**Feature**: Register and Login Authentication  
**Context**: Multi-tenant authentication for echoforge platform

## Research Questions Resolved

### 1. JWT Secret Configuration Source
**Question**: Where should JWT secret be configured - environment variable JWT_SECRET or config.yaml?  
**Decision**: Environment variable `JWT_SECRET`  
**Rationale**: 
- Follows 12-factor app principles for configuration
- Prevents accidental commit of secrets to version control
- Enables different secrets per deployment environment
- Standard practice in Go authentication libraries
- Compatible with Docker container deployments
**Alternatives considered**: 
- config.yaml: Rejected due to security risk of committing secrets
- Dynamic generation: Rejected due to stateless JWT requirement across instances

### 2. Rate Limiting Implementation Strategy
**Question**: Should rate limiting be implemented in usecase layer or deferred to Gin middleware?  
**Decision**: Defer to Gin middleware layer  
**Rationale**: 
- Separation of concerns: usecases handle business logic, middleware handles cross-cutting concerns  
- Rate limiting is an infrastructure concern, not business logic
- Gin middleware provides better performance and flexibility
- Allows different rate limits per endpoint type
- Compatible with future API gateway deployment
- Follows hexagonal architecture principles (adapters handle external concerns)
**Alternatives considered**: 
- Usecase layer: Rejected as it violates single responsibility and complicates testing
- Database-based: Rejected due to performance overhead for every auth request

### 3. Password Validation Policy
**Question**: What constitutes OWASP-compliant strong password policy?  
**Decision**: Minimum 8 characters with go-playground/validator/v10 custom validation  
**Rationale**: 
- OWASP recommends minimum 8 characters over complex character requirements
- Avoid password composition rules that reduce usability
- Focus on length over complexity for better security/UX balance
- bcrypt handles hashing regardless of password strength
**Implementation**: Custom validator function checking minimum length and basic strength indicators

### 4. JWT Token Claims Structure
**Question**: What claims should be included in JWT tokens?  
**Decision**: Standard claims (sub, exp, iat) + custom claims (site_id)  
**Rationale**: 
- `sub` (subject): User ID for identification
- `exp` (expiration): 24 hours as specified in requirements
- `iat` (issued at): For token freshness validation
- `site_id`: Critical for multi-tenant authorization
- Minimal claims reduce token size and avoid sensitive data exposure
**Alternatives considered**: Including email/username rejected due to token size and potential data exposure

### 5. Error Response Strategy
**Question**: How should authentication errors be structured to prevent information disclosure?  
**Decision**: Generic "authentication failed" for login, specific validation errors for registration  
**Rationale**: 
- Login failures should not reveal whether email exists (security)  
- Registration validation should provide specific feedback (UX)
- Structured errors using errors.Join for validation failures
- Consistent error format across usecases
**Implementation**: Different error handling strategies per usecase based on security requirements

### 6. Dependency Integration
**Question**: How to integrate github.com/golang-jwt/jwt/v5 with existing architecture?  
**Decision**: Add JWT utilities to existing `pkg/auth/jwt.go` with generator/validator functions  
**Rationale**: 
- Leverages existing auth package structure
- Centralizes JWT operations for reusability
- Maintains separation from business logic
- Testable utility functions
- Compatible with dependency injection pattern

## Technology Research

### bcrypt Cost Factor
**Best Practice**: Use bcrypt.DefaultCost (currently 10) or higher for production  
**Recommendation**: Use cost=12 for improved security with acceptable performance  
**Rationale**: Balances security against authentication latency requirements (<2 seconds)

### Context Handling
**Best Practice**: All usecase methods should accept context.Context as first parameter  
**Implementation**: Support cancellation and timeout for database operations  
**Rationale**: Enables graceful shutdown and request timeout handling

### Input Validation
**Best Practice**: Use go-playground/validator/v10 struct tags + custom validation functions  
**Implementation**: Validate at usecase input boundaries before business logic  
**Rationale**: Fail fast, clear error messages, standardized validation approach

## Architecture Decisions

### Usecase Organization
**Structure**: `internal/usecase/user/` with separate files per usecase  
**Rationale**: 
- Follows single responsibility principle
- Enables parallel development
- Clear testing boundaries
- Future-proof for additional auth features

### Dependency Injection
**Approach**: Constructor injection with interfaces  
**Pattern**: `NewRegisterUsecase(repo domain.UserRepository, jwtSecret string)`  
**Rationale**: 
- Testable with mock repositories
- Compatible with Wire dependency injection
- Clear dependency declarations
- Follows hexagonal architecture ports pattern

### Testing Strategy
**Approach**: Comprehensive unit tests with mock repository + integration tests  
**Coverage Target**: >80% as per constitutional requirements  
**Test Types**: 
- Unit tests: Business logic with mocked dependencies
- Contract tests: Input validation and error handling
- Integration tests: End-to-end with real repository (future HTTP layer)

## Security Considerations

### OWASP Compliance
- Input validation prevents injection attacks
- Generic error messages prevent user enumeration
- Secure password hashing with bcrypt
- JWT tokens with appropriate expiration
- Rate limiting deferred to infrastructure layer

### Multi-Tenant Security
- All operations scoped by siteID
- No cross-tenant data access possible
- JWT tokens include site context
- Repository layer enforces tenant isolation

## Performance Considerations

### Concurrency
- Stateless usecases support concurrent execution
- bcrypt hashing may benefit from goroutine pools for high load
- JWT generation/validation is CPU-intensive but cacheable

### Scalability
- Horizontal scaling compatible (stateless design)
- JWT tokens eliminate session storage requirements
- Database connection pooling handled by GORM

---

**Research Status**: COMPLETE  
**All NEEDS CLARIFICATION Resolved**: YES  
**Ready for Phase 1 Design**: YES