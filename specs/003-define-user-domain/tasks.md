# Tasks: User Domain Entity and Repository

**Input**: Design documents from `/specs/003-define-user-domain/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory ✓
   → Tech stack: Go 1.25+, GORM v1.26+, Testify, UUID, bcrypt
   → Structure: Hexagonal architecture (internal/domain, adapters/persistence)
2. Load design documents ✓:
   → data-model.md: User entity with validation
   → contracts/user_repository.md: Repository interface specification
   → quickstart.md: Implementation examples and test scenarios
3. Generate tasks by category ✓:
   → Domain: Pure User entity with validation in internal/domain
   → Repository: Interface + GORM implementation + mock for testing
   → Database: Migration files for PostgreSQL with multi-tenant schema
   → Tests: TDD approach with unit, integration, and contract tests
   → Use cases: Business logic layer using repository interface  
4. Apply task rules ✓:
   → Different files = mark [P] for parallel execution
   → Tests before implementation (TDD with 80%+ coverage)
   → Multi-tenant isolation via site_id in all operations
   → GORM v1.26+ with golang-migrate for all persistence
5. Number tasks sequentially (T001-T018) ✓
6. Dependencies: Domain → Repository → Use Cases → Integration ✓
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- File paths are absolute from repository root

## Phase 3.1: Foundation Setup
- [x] **T001** [P] Create database migration 002_create_users_table.up.sql in `C:\Users\Nhan\go\src\echoforge\migrations\002_create_users_table.up.sql` with User table schema (UUID primary key, site_id, email, password_hash, timestamps, unique constraint on site_id+email)
- [x] **T002** [P] Create database migration 002_create_users_table.down.sql in `C:\Users\Nhan\go\src\echoforge\migrations\002_create_users_table.down.sql` to drop users table

## Phase 3.2: Domain Layer Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] **T003** [P] Create User entity unit tests in `C:\Users\Nhan\go\src\echoforge\tests\user_domain_test.go` covering NewUser creation, validation rules (email format, password hash length, required fields), and IsValid method
- [x] **T004** [P] Create UserRepository contract tests in `C:\Users\Nhan\go\src\echoforge\tests\user_repository_contract_test.go` testing Create and FindByEmail methods with mock implementation, context cancellation, and error conditions

## Phase 3.3: Domain Layer Implementation (Tests must be failing first)
- [x] **T005** Create User domain entity in `C:\Users\Nhan\go\src\echoforge\internal\domain\user.go` with UUID ID, SiteID, Email, PasswordHash, timestamps, validation methods (NewUser, Validate, IsValid), email regex validation, password hash length check, and all domain errors
- [x] **T006** Add UserRepository interface to `C:\Users\Nhan\go\src\echoforge\internal\domain\user.go` with Create(ctx, user) and FindByEmail(ctx, siteID, email) methods, proper error types (ErrUserAlreadyExists, ErrRepositoryFailure)

## Phase 3.4: Repository Implementation Tests
- [x] **T007** [P] Create UserRepository integration tests in `C:\Users\Nhan\go\src\echoforge\tests\user_repository_test.go` using real PostgreSQL database, testing Create with duplicate email within same site, FindByEmail with site isolation, context timeout handling, and database transaction rollback
- [x] **T008** [P] Create mock UserRepository implementation in `C:\Users\Nhan\go\src\echoforge\tests\mock_user_repository.go` for use case testing with in-memory storage, call tracking, and reset functionality

## Phase 3.5: Repository Implementation 
- [x] **T009** Create GORM UserRepository implementation in `C:\Users\Nhan\go\src\echoforge\adapters\persistence\user_repository.go` with GORM model mapping, Create method with unique constraint handling, FindByEmail with site_id filtering, proper error wrapping, and context support

## Phase 3.6: Use Case Layer Tests
- [x] **T010** [P] Create User use case unit tests in `C:\Users\Nhan\go\src\echoforge\tests\user_usecase_test.go` using mock repository to test user creation workflow, email uniqueness validation within site, password hash requirements, and error handling scenarios

## Phase 3.7: Use Case Implementation
- [x] **T011** Create User use case in `C:\Users\Nhan\go\src\echoforge\internal\usecase\user_usecase.go` with CreateUser business logic using repository interface, email validation before persistence, proper error handling and logging, and context propagation

## Phase 3.8: Integration & Validation
- [ ] **T012** [P] Run all domain unit tests and verify 80%+ code coverage for `internal/domain/user.go` using `go test -cover`
- [ ] **T013** [P] Run repository integration tests against real PostgreSQL database with proper setup/teardown and migration execution
- [ ] **T014** [P] Validate multi-tenant isolation by creating users with same email in different sites and ensuring FindByEmail returns correct site-specific user
- [ ] **T015** [P] Execute quickstart.md validation by running all example code snippets and verifying they work as documented

## Phase 3.9: Performance & Security
- [ ] **T016** [P] Create database indexes on users table (site_id, site_id+email composite) and verify query performance for FindByEmail operation
- [ ] **T017** [P] Validate password hash security by testing bcrypt hash length validation and ensuring no plaintext passwords are accepted
- [ ] **T018** [P] Run concurrent user creation test to verify 1000+ operations per site capability and ensure no race conditions in unique email constraint

## Dependencies
```
Foundation: T001, T002 (parallel)
↓
Domain Tests: T003, T004 (parallel) 
↓
Domain Implementation: T005 → T006 (sequential - same file)
↓
Repository Tests: T007, T008 (parallel)
↓  
Repository Implementation: T009
↓
Use Case Tests: T010
↓
Use Case Implementation: T011  
↓
Validation: T012, T013, T014, T015 (parallel)
↓
Performance: T016, T017, T018 (parallel)
```

## Parallel Execution Examples

### Phase 3.1 (Foundation)
```bash
# Run migration tasks in parallel:
# Task T001: Create up migration
# Task T002: Create down migration
```

### Phase 3.2 (Domain Tests)  
```bash
# Run test creation in parallel:
# Task T003: Domain entity unit tests
# Task T004: Repository contract tests
```

### Phase 3.4 (Repository Tests)
```bash
# Run repository test tasks in parallel:
# Task T007: Integration tests with real DB
# Task T008: Mock repository for unit tests
```

### Phase 3.8 (Validation)
```bash
# Run validation tasks in parallel:
# Task T012: Coverage verification
# Task T013: Integration test execution  
# Task T014: Multi-tenant isolation test
# Task T015: Quickstart validation
```

### Phase 3.9 (Performance)
```bash
# Run performance tasks in parallel:
# Task T016: Database indexing
# Task T017: Security validation
# Task T018: Concurrency testing
```

## Quality Gates
- All tests in Phase 3.2 must **FAIL** before starting Phase 3.3
- Test coverage must exceed **80%** for domain layer (verified in T012)
- All repository operations must include **site_id** filtering (verified in T014)
- Password validation must enforce **60+ character** bcrypt hashes (verified in T017)
- System must handle **1000+ concurrent** operations per site (verified in T018)

## File Structure Target
```
internal/
├── domain/
│   └── user.go                    # User entity + UserRepository interface (T005, T006)
└── usecase/
    └── user_usecase.go            # Business logic using repository (T011)

adapters/
└── persistence/
    └── user_repository.go         # GORM implementation (T009)

tests/
├── user_domain_test.go            # Domain unit tests (T003)
├── user_repository_contract_test.go # Repository contract tests (T004)
├── user_repository_test.go        # Repository integration tests (T007)
├── user_usecase_test.go           # Use case unit tests (T010)
└── mock_user_repository.go        # Mock for testing (T008)

migrations/
├── 002_create_users_table.up.sql   # Schema creation (T001)
└── 002_create_users_table.down.sql # Schema rollback (T002)
```

## Constitutional Compliance Checklist
- [x] Hexagonal architecture maintained (domain/usecase/adapters separation)
- [x] GORM v1.26+ used for all persistence operations
- [x] TDD enforced with tests written before implementation
- [x] Multi-site tenant isolation via site_id in all operations
- [x] 80%+ test coverage requirement included
- [x] Security validation for bcrypt password hashing
- [x] Performance testing for 1000+ concurrent users per site
- [x] Lean MVP approach (focused on core User entity only)