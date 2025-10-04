# Tasks: Database Connection and Authentication APIs

**Input**: Design documents from `/specs/006-i-already-create/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extract: Go 1.25+, GORM v1.26+, Gin v1.10+, PostgreSQL, JWT auth
2. Load design documents:
   → data-model.md: User entity → model tasks
   → contracts/auth-api.md: 5 endpoints → contract test tasks
   → research.md: bcrypt+JWT+rate limiting → security tasks
   → quickstart.md: Database setup → integration tasks
3. Generate tasks by category:
   → Setup: Go project, dependencies, database migrations
   → Tests: TDD with Testify (80%+ coverage), contract tests per endpoint
   → Core: User domain entity, authentication use cases, JWT utilities
   → Adapters: GORM user repository, Gin HTTP handlers, rate limiting middleware
   → Security: bcrypt password hashing, JWT token management, input validation
   → Integration: database connection, server startup, configuration
   → Polish: health checks, error handling, documentation updates
4. Apply constitutional requirements:
   → Hexagonal architecture: domain in internal/domain, adapters separated
   → GORM v1.26+ with golang-migrate for all persistence
   → Gin v1.10+ with /api/v1/ versioned endpoints
   → bcrypt+JWT auth with unique email per site, rate limiting
   → TDD approach with 80%+ test coverage using Testify
   → Performance target: 1000+ concurrent users, <500ms response times
5. Task ordering: Tests before implementation (TDD)
6. Parallel execution: Different files marked [P]
7. Dependencies: Models before services, services before endpoints
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- File paths follow modular monolith hexagonal architecture

## Path Conventions (Modular Monolith)
```
internal/domain/        # Pure entities, interfaces
internal/usecase/       # Business logic with DI
adapters/http/          # Gin HTTP handlers  
adapters/persistence/   # GORM repositories
cmd/server/            # Main application entry
pkg/auth/              # JWT, bcrypt utilities
pkg/common/            # Shared utilities
tests/                 # TDD tests (80%+ coverage)
migrations/            # golang-migrate DB schemas
configs/               # Configuration files
```

## Phase 3.1: Setup
- [x] T001 Initialize Go 1.25+ project with go.mod and required dependencies (GORM v1.26+, Gin v1.10+, JWT v5.3+, bcrypt, Testify v1.11+, Wire, golang-migrate)
- [x] T002 [P] Create modular monolith directory structure (internal/domain, internal/usecase, adapters/, cmd/server/, pkg/, tests/, migrations/)
- [x] T003 [P] Configure PostgreSQL database connection in configs/config.yaml with Viper configuration management
- [x] T004 [P] Set up golang-migrate for database migrations with PostgreSQL driver

## Phase 3.2: Tests First (TDD) ✅ COMPLETE
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T005 [P] Contract test POST /api/v1/auth/register in tests/contract/auth_register_test.go
- [x] T006 [P] Contract test POST /api/v1/auth/login in tests/contract/auth_login_test.go  
- [x] T007 [P] Contract test POST /api/v1/auth/logout in tests/contract/auth_logout_test.go
- [x] T008 [P] Contract test GET /api/v1/auth/profile in tests/contract/auth_profile_test.go
- [x] T009 [P] Contract test GET /api/v1/health in tests/contract/health_test.go
- [x] T010 [P] Integration test user registration flow in tests/integration/user_registration_test.go
- [x] T011 [P] Integration test authentication flow in tests/integration/auth_flow_test.go
- [x] T012 [P] Integration test rate limiting in tests/integration/rate_limiting_test.go

## Phase 3.3: Core Domain ✅ COMPLETE  
- [x] T013 [P] User entity with validation in internal/domain/user.go (UUID, email, password hash, timestamps)
- [x] T014 [P] User repository interface in internal/domain/user_repository.go
- [x] T015 [P] Authentication domain service interface in internal/domain/auth_service.go
- [x] T016 [P] JWT utilities (generate, validate, blacklist) in pkg/auth/jwt.go
- [x] T017 [P] Password utilities (hash, verify with bcrypt cost 12) in pkg/auth/password.go

## Phase 3.4: Use Cases & Business Logic ✅ COMPLETE
- [x] T018 User registration use case in internal/usecase/user_registration.go (email uniqueness, password validation)
- [x] T019 User authentication use case in internal/usecase/user_authentication.go (login, token generation)
- [x] T020 User logout use case in internal/usecase/user_logout.go (token blacklisting)
- [x] T021 Get user profile use case in internal/usecase/get_user_profile.go (JWT validation)

## Phase 3.5: Persistence Adapters (GORM) ✅ COMPLETE
- [x] T022 [P] Create user migration 001_create_users_table.up.sql in migrations/ (UUID primary key, unique email index)
- [x] T023 [P] Create auth blacklist migration 002_create_auth_blacklist.up.sql in migrations/ (for JWT logout)
- [x] T024 User repository implementation with GORM in adapters/persistence/user_repository.go
- [x] T025 Database connection and auto-migration in adapters/persistence/database.go

## Phase 3.6: HTTP Adapters (Gin) ✅ COMPLETE
- [x] T026 [P] POST /api/v1/auth/register handler in adapters/http/auth_handler.go
- [x] T027 [P] POST /api/v1/auth/login handler in adapters/http/auth_handler.go  
- [x] T028 [P] POST /api/v1/auth/logout handler in adapters/http/auth_handler.go
- [x] T029 [P] GET /api/v1/auth/profile handler in adapters/http/auth_handler.go
- [x] T030 [P] GET /api/v1/health handler in adapters/http/health_handler.go
- [x] T031 JWT authentication middleware in adapters/http/middleware/auth_middleware.go
- [x] T032 [P] Rate limiting middleware (5/min per IP) in adapters/http/middleware/rate_limit_middleware.go
- [x] T033 [P] Input validation middleware with go-playground/validator/v10 in adapters/http/middleware/validation_middleware.go

## Phase 3.7: Application Integration ✅ COMPLETE
- [x] T034 Wire dependency injection setup in cmd/server/wire.go (connect all adapters)
- [x] T035 Gin router setup with versioned endpoints in cmd/server/router.go
- [x] T036 Server configuration and startup in cmd/server/main.go (Viper config, Zap logging)
- [x] T037 [P] Configuration validation and environment setup in cmd/server/config.go

## Phase 3.8: Security & Performance ✅ COMPLETE
- [x] T038 [P] OWASP Top 10 compliance validation (input sanitization, security headers)
- [x] T039 [P] Performance optimization (database connection pooling, concurrent request handling)
- [x] T040 [P] Security logging with Zap (authentication events, failed attempts)
- [x] T041 [P] Error handling with generic messages (avoid information disclosure)

## Phase 3.9: Polish & Documentation
- [x] T042 [P] Unit tests for domain entities in tests/unit/domain/user_test.go
- [x] T043 [P] Unit tests for JWT utilities in tests/unit/auth/jwt_test.go
- [x] T044 [P] Unit tests for use cases in tests/unit/usecase/
- [x] T045 [P] Performance tests targeting <500ms response times in tests/performance/
- [x] T046 [P] Update README.md with setup and usage instructions
- [x] T047 [P] Update API documentation based on implemented endpoints
- [x] T048: Code coverage analysis and optimization ✅

## Dependencies
- **Setup Phase** (T001-T004) must complete before all other phases
- **Test Phase** (T005-T012) must complete before implementation phases
- **Core Domain** (T013-T017) must complete before use cases
- **Use Cases** (T018-T021) must complete before adapters
- **Migrations** (T022-T023) must complete before persistence adapters
- **All adapters** must complete before integration phase
- **Integration** (T034-T037) required before security and performance
- **Polish** (T042-T048) can run after core implementation is complete

## Parallel Execution Examples

### Phase 3.2: Contract Tests (All Parallel)
```bash
# Launch all contract tests simultaneously:
Task: "Contract test POST /api/v1/auth/register in tests/contract/auth_register_test.go"
Task: "Contract test POST /api/v1/auth/login in tests/contract/auth_login_test.go"
Task: "Contract test POST /api/v1/auth/logout in tests/contract/auth_logout_test.go"
Task: "Contract test GET /api/v1/auth/profile in tests/contract/auth_profile_test.go"
Task: "Contract test GET /api/v1/health in tests/contract/health_test.go"
```

### Phase 3.3: Core Domain (All Parallel)
```bash
# Launch all domain entities simultaneously:
Task: "User entity with validation in internal/domain/user.go"
Task: "User repository interface in internal/domain/user_repository.go"
Task: "Authentication domain service interface in internal/domain/auth_service.go"
Task: "JWT utilities in pkg/auth/jwt.go"
Task: "Password utilities in pkg/auth/password.go"
```

### Phase 3.6: HTTP Handlers (Most Parallel)
```bash
# Launch independent handlers simultaneously:
Task: "POST /api/v1/auth/register handler in adapters/http/auth_handler.go"
Task: "GET /api/v1/health handler in adapters/http/health_handler.go"
Task: "Rate limiting middleware in adapters/http/middleware/rate_limit_middleware.go"
Task: "Input validation middleware in adapters/http/middleware/validation_middleware.go"
```

## Constitutional Compliance Checklist
- [x] Modular monolith with hexagonal architecture (internal/domain, adapters/)
- [x] GORM v1.26+ with golang-migrate for all PostgreSQL persistence
- [x] Gin v1.10+ with versioned /api/v1/ endpoints
- [x] TDD with Testify, targeting 80%+ test coverage
- [x] Clone-and-extend model (separate DB per site via config)
- [x] bcrypt + JWT authentication with email uniqueness per site
- [x] Rate limiting and OWASP Top 10 security compliance
- [x] Performance target: 1000+ concurrent users, <500ms response
- [x] Dependency injection with Wire for clean architecture
- [x] Structured logging with Zap for observability
- [x] Configuration management with Viper for multi-site support

## Success Criteria
1. **All 48 tasks completed** with proper dependency ordering
2. **80%+ test coverage** achieved using Testify framework
3. **All 5 API endpoints** implemented and tested (register, login, logout, profile, health)
4. **Authentication flow** working end-to-end with JWT tokens
5. **Rate limiting** preventing brute force attacks (5/min per IP)
6. **Database integration** with PostgreSQL using GORM and migrations
7. **Performance target** met (<500ms response times, 1000+ concurrent users)
8. **Security compliance** with OWASP Top 10 and bcrypt password hashing
9. **Configuration system** supporting multi-site deployment model
10. **Documentation** updated with API usage and deployment instructions

**Tasks Status**: Ready for execution with TDD approach and constitutional compliance ✅