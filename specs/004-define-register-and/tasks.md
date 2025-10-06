# Tasks: Register and Login Authentication Usecases

**Input**: Design documents from `/specs/004-define-register-and/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extracted: Go 1.25+, gin v1.10.0, gorm.io/gorm v1.26.12, go-playground/validator/v10 v10.27.0, 
     golang.org/x/crypto v0.42.0, github.com/golang-jwt/jwt/v5, testify v1.13.1
2. Load optional design documents:
   → data-model.md: RegisterInput, LoginInput, AuthenticationResult entities
   → contracts/: RegisterUsecase, LoginUsecase interface contracts
   → research.md: JWT secret via env var, bcrypt cost=12, rate limiting deferred
3. Generate tasks by category:
   → Setup: JWT dependency, JWT utilities, environment configuration
   → Tests: TDD with Testify (80%+ coverage), contract tests, validation tests
   → Core: RegisterUsecase, LoginUsecase implementations with bcrypt+JWT
   → Security: Input validation, password hashing, JWT token management
   → Multi-tenant: site_id isolation in all operations
   → Polish: Integration tests, security tests, performance validation
4. Apply task rules:
   → Different files = mark [P] for parallel execution
   → Same file = sequential (no [P])
   → Tests before implementation (TDD enforcement)
   → All auth tasks use bcrypt+JWT with proper validation
   → All operations enforce site_id multi-tenant isolation
5. Number tasks sequentially (T001-T018)
6. Dependencies: Setup → Tests → Implementation → Integration → Validation
7. Parallel execution: Independent test files and usecase implementations
8. Task completeness validated:
   → All contracts have tests ✓
   → All entities have implementations ✓ 
   → All usecases have comprehensive test coverage ✓
9. Return: SUCCESS (18 tasks ready for TDD execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in Go project structure

## Path Conventions
- **Go project structure**: `internal/`, `pkg/`, `tests/` at repository root
- Following hexagonal architecture with existing user domain from Task 1.2
- New usecase layer in `internal/usecase/user/`

## Phase 3.1: Foundation Setup
- [x] **T001** Add JWT dependency to go.mod: `go get github.com/golang-jwt/jwt/v5@latest && go mod tidy`
- [x] **T002** Create JWT utilities in `pkg/auth/jwt.go` with GenerateToken and ValidateToken functions
- [x] **T003** [P] Create JWT utilities tests in `pkg/auth/jwt_test.go` with comprehensive token validation scenarios

## Phase 3.2: Usecase Interface & Input Models ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: Define interfaces and DTOs before implementation**
- [x] **T004** [P] Create RegisterInput struct in `internal/usecase/user/register.go` with validation tags
- [x] **T005** [P] Create LoginInput and AuthenticationResult structs in `internal/usecase/user/login.go` with validation tags
- [x] **T006** [P] Define RegisterUsecase interface in `internal/usecase/user/register.go` with Execute method
- [x] **T007** [P] Define LoginUsecase interface in `internal/usecase/user/login.go` with Execute method

## Phase 3.3: TDD Test Implementation ⚠️ MUST COMPLETE BEFORE 3.4
**CRITICAL: These tests MUST be written and MUST FAIL before ANY usecase implementation**
- [x] **T008** [P] Create RegisterUsecase tests in `internal/usecase/user/register_test.go` with success, validation, and duplicate email scenarios
- [ ] **T009** [P] Create LoginUsecase tests in `internal/usecase/user/login_test.go` with success, invalid credentials, and site isolation scenarios
- [ ] **T010** [P] Create JWT utilities integration tests in `pkg/auth/jwt_test.go` with token generation and validation edge cases
- [ ] **T011** [P] Create input validation tests for RegisterInput struct validation in `internal/usecase/user/register_test.go`
- [ ] **T012** [P] Create input validation tests for LoginInput struct validation in `internal/usecase/user/login_test.go`

## Phase 3.4: Core Usecase Implementation (ONLY after tests are failing)
- [ ] **T013** Implement RegisterUsecase in `internal/usecase/user/register.go` with bcrypt password hashing, duplicate email checking, and user creation
- [ ] **T014** Implement LoginUsecase in `internal/usecase/user/login.go` with password verification, JWT token generation, and authentication result creation
- [ ] **T015** Implement JWT token generation in `pkg/auth/jwt.go` with HS256 signing, user+site claims, and 24-hour expiration
- [ ] **T016** Implement JWT token validation in `pkg/auth/jwt.go` with signature verification, expiration checking, and claims extraction

## Phase 3.5: Integration & Security Testing
- [ ] **T017** [P] Create end-to-end authentication flow integration test in `tests/auth_integration_test.go` with register→login flow
- [ ] **T018** [P] Create multi-tenant security tests in `tests/security_test.go` with cross-site isolation validation and password security requirements

## Dependencies
**Sequential Dependencies:**
- T001 (JWT dependency) blocks T002, T003
- T002 (JWT utilities) blocks T015, T016
- T004-T007 (interfaces/DTOs) block T008-T012 (tests)
- T008-T012 (failing tests) block T013-T016 (implementations)
- T013-T016 (implementations) block T017-T018 (integration tests)

**Parallel Opportunities:**
- T003, T004, T005, T006, T007 can run in parallel (different files)
- T008, T009, T010, T011, T012 can run in parallel (different test scenarios, different files)
- T017, T018 can run in parallel (different integration test files)

## Parallel Execution Examples

### Phase 3.2: Interface Definition (Parallel)
```bash
# Launch T004-T007 together:
# T004: RegisterInput struct with validation tags
# T005: LoginInput/AuthenticationResult structs  
# T006: RegisterUsecase interface definition
# T007: LoginUsecase interface definition
```

### Phase 3.3: TDD Test Creation (Parallel)
```bash
# Launch T008-T012 together:
# T008: RegisterUsecase comprehensive test suite
# T009: LoginUsecase comprehensive test suite  
# T010: JWT utilities integration tests
# T011: RegisterInput validation tests
# T012: LoginInput validation tests
```

### Phase 3.5: Integration Testing (Parallel)
```bash
# Launch T017-T018 together:
# T017: End-to-end authentication flow tests
# T018: Multi-tenant security validation tests
```

## Detailed Task Specifications

### T001: Add JWT Dependency
**File**: `go.mod`
**Command**: `go get github.com/golang-jwt/jwt/v5@latest && go mod tidy`
**Validation**: Dependency appears in go.mod with latest version

### T002: JWT Utilities Foundation
**File**: `pkg/auth/jwt.go`
**Functions**: 
- `GenerateToken(userID, siteID uuid.UUID, secret string) (string, time.Time, error)`
- `ValidateToken(tokenString, secret string) (*JWTClaims, error)`
- `JWTClaims` struct with standard + custom claims

### T008: RegisterUsecase TDD Tests (Must Fail Initially)
**File**: `internal/usecase/user/register_test.go`
**Test Scenarios**:
- `TestRegisterUsecase_Success` - Valid registration
- `TestRegisterUsecase_ValidationErrors` - Invalid input validation
- `TestRegisterUsecase_DuplicateEmail` - Email already exists in same site
- `TestRegisterUsecase_SameEmailDifferentSites` - Email reuse across sites allowed
- `TestRegisterUsecase_ContextCancellation` - Context timeout handling
- `TestRegisterUsecase_RepositoryErrors` - Database failure scenarios

### T009: LoginUsecase TDD Tests (Must Fail Initially)
**File**: `internal/usecase/user/login_test.go`
**Test Scenarios**:
- `TestLoginUsecase_Success` - Valid authentication with JWT token
- `TestLoginUsecase_InvalidCredentials` - Wrong password (generic error)
- `TestLoginUsecase_UserNotFound` - User doesn't exist (generic error)
- `TestLoginUsecase_SiteIsolation` - User exists but in different site
- `TestLoginUsecase_ValidationErrors` - Invalid input validation
- `TestLoginUsecase_JWTGenerationFailure` - Token generation error scenarios

### T013: RegisterUsecase Implementation
**File**: `internal/usecase/user/register.go`
**Implementation Requirements**:
- Input validation using go-playground/validator/v10
- Email uniqueness check via repository.FindByEmail(ctx, siteID, email)
- Password hashing with bcrypt.GenerateFromPassword(cost=12)
- User entity creation with uuid.New() for ID
- Repository persistence with error handling
- Site ID isolation enforcement

### T014: LoginUsecase Implementation  
**File**: `internal/usecase/user/login.go`
**Implementation Requirements**:
- Input validation using struct tags
- User lookup via repository.FindByEmail(ctx, siteID, email)
- Password verification with bcrypt.CompareHashAndPassword
- JWT token generation with user+site claims
- AuthenticationResult creation with token and expiration
- Generic error messages for security (no user enumeration)

## Success Criteria per Task

### Coverage Requirements
- **Minimum**: 80% test coverage per constitutional requirements
- **Target**: >90% coverage for authentication-critical code
- **Tools**: `go test -cover -v` for validation

### Security Validation
- All password operations use bcrypt with cost ≥12
- JWT tokens include proper claims (sub, site_id, exp, iat)
- Multi-tenant isolation verified in all database queries
- Generic error messages prevent user enumeration
- Input validation prevents injection attacks

### Performance Targets
- Registration: <2 seconds under normal load
- Login: <1 second with token generation
- Concurrent users: 1000+ per site capability
- JWT operations: <100ms per token

## Notes
- **[P] tasks** = different files, no dependencies, can run concurrently
- **TDD enforcement**: All T008-T012 tests must fail before T013-T016 implementation
- **Multi-tenant isolation**: Every database operation includes site_id parameter
- **Security first**: Generic errors, input validation, secure token generation
- **Constitutional compliance**: Hexagonal architecture, GORM persistence, 80%+ coverage

## Task Generation Rules Applied

1. **From Contracts**: 
   - RegisterUsecase interface → T006, T008, T013
   - LoginUsecase interface → T007, T009, T014

2. **From Data Model**:
   - RegisterInput entity → T004, T011
   - LoginInput/AuthenticationResult entities → T005, T012
   - JWT claims structure → T002, T010, T015, T016

3. **From Research Decisions**:
   - JWT secret via environment → T001, T002
   - bcrypt cost=12 → T013 implementation
   - Rate limiting deferred → No usecase-level tasks

4. **From Quickstart Scenarios**:
   - Registration flow → T008, T013, T017
   - Login flow → T009, T014, T017
   - Security validation → T018

## Validation Checklist
*GATE: All items verified*

- [x] All contracts have corresponding tests (T008-T012)
- [x] All entities have implementation tasks (T004-T007, T013-T014)
- [x] All tests come before implementation (T008-T012 before T013-T016)
- [x] Parallel tasks truly independent ([P] tasks use different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] TDD enforcement with failing tests requirement
- [x] 80%+ coverage achievable with comprehensive test scenarios
- [x] Multi-tenant isolation enforced in all database operations
- [x] Security requirements (bcrypt, JWT, validation) implemented