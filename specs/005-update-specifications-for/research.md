# Research: Update User Domain and Authentication for Clone-and-Extend Model

**Date**: October 4, 2025 | **Spec**: [spec.md](./spec.md)

## Current State Analysis

### 1. Domain Layer (`internal/domain/user.go`)
**Current Architecture**: Multi-tenant with `site_id` isolation
- User struct contains `SiteID uuid.UUID` field
- `NewUser()` requires `siteID` parameter
- `Validate()` enforces `SiteID` as required field
- `UserRepository.FindByEmail()` requires `siteID` parameter for tenant isolation

### 2. Use Cases (`internal/usecase/user/`)
**Current Implementation**: Site-scoped operations
- `RegisterInput` struct contains `SiteID uuid.UUID` field with validation
- `LoginInput` struct contains `SiteID uuid.UUID` field with validation
- Both use cases pass `siteID` to repository methods
- JWT generation includes both `userID` and `siteID` claims

### 3. Authentication (`pkg/auth/jwt.go`)
**Current JWT Claims**: Multi-tenant aware
- `JWTClaims` contains both `UserID` and `SiteID` fields
- `GenerateToken()` requires both `userID` and `siteID` parameters
- Token validation returns both user and site context

## Technical Decisions

### 1. Architecture Alignment
**Decision**: Maintain hexagonal architecture while removing site_id coupling
- **Rationale**: Constitution v1.2.0 mandates clone-and-extend model where each site runs separate database instance
- **Impact**: Eliminates need for runtime tenant isolation via `site_id` queries
- **Benefit**: Simplifies domain logic while maintaining performance and security

### 2. Database Strategy
**Decision**: Single-site per database instance
- **Rationale**: Clone-and-extend model provides natural tenant isolation at database level
- **Impact**: Removes `site_id` from all database queries and constraints
- **Benefit**: Improved query performance, simplified schema, natural data isolation

### 3. JWT Token Strategy
**Decision**: User-only claims without site context
- **Rationale**: Each site clone operates independently with its own user base
- **Impact**: JWT tokens contain only user ID, removing site_id claim
- **Benefit**: Simplified token validation, reduced token size, clearer security model

### 4. Backward Compatibility
**Decision**: Breaking change acceptable for architectural refactoring
- **Rationale**: This is a major architectural shift to clone-and-extend model
- **Impact**: Existing multi-tenant deployments need migration to separate instances
- **Benefit**: Long-term maintainability and performance improvements

## Implementation Strategy

### Phase 1: Domain Entity Updates
1. Remove `SiteID` field from `User` struct
2. Update `NewUser()` constructor to remove `siteID` parameter
3. Update `Validate()` method to remove `SiteID` validation
4. Update `UserRepository` interface to remove `siteID` parameters

### Phase 2: Use Case Updates
1. Remove `SiteID` from `RegisterInput` and `LoginInput` structs
2. Update validation tags to remove `site_id` requirements
3. Update use case implementations to remove site-scoped operations
4. Update error messages to remove site context

### Phase 3: Authentication Updates
1. Remove `SiteID` field from `JWTClaims` struct
2. Update `GenerateToken()` to accept only `userID` parameter
3. Update token validation to handle user-only claims
4. Maintain token expiration and security standards

### Phase 4: Test Updates
1. Update all domain tests to remove `siteID` parameters
2. Update use case tests to remove site isolation scenarios
3. Update JWT tests to handle simplified claims structure
4. Maintain 80%+ test coverage requirement

## Risk Assessment

### Low Risk
- **Performance**: Removing `site_id` queries improves database performance
- **Security**: Database-level isolation stronger than application-level isolation
- **Maintainability**: Simplified code with fewer parameters and validations

### Medium Risk
- **Migration Complexity**: Existing multi-tenant data needs careful migration planning
- **Configuration Changes**: Sites need separate configuration and deployment

### Mitigation Strategies
- Provide clear migration documentation in `quickstart.md`
- Maintain comprehensive test coverage during refactoring
- Use TDD approach to validate each change

## Dependencies

### Internal Dependencies
- Domain entities: `internal/domain/user.go`
- Use cases: `internal/usecase/user/register.go`, `internal/usecase/user/login.go`
- Authentication: `pkg/auth/jwt.go`
- Tests: All user-related test files

### External Dependencies
- GORM v1.26+: No changes needed (benefits from simplified queries)
- go-playground/validator/v10: Update validation tags
- golang-jwt/jwt/v5: No changes needed (fewer claims)
- bcrypt: No changes needed
- testify: No changes needed

## Success Criteria
1. All `site_id` references removed from domain, use cases, and authentication
2. User registration and login work without site context
3. JWT tokens contain only user ID claims
4. All tests pass with 80%+ coverage
5. No performance degradation (actually improved due to simpler queries)
6. Constitution v1.2.0 compliance verified

## Next Steps
Proceed to Phase 1: Design & Contracts to create detailed interface definitions and data models.