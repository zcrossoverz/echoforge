# Tasks: Update User Domain and Authentication for Clone-and-Extend Model

**Input**: Design documents from `/specs/005-update-specifications-for/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extract: Go 1.25+, Gin v1.10+, GORM v1.26+, JWT auth, hexagonal architecture
   → Target: Remove site_id from User domain and authentication system
2. Load design documents:
   → data-model.md: User entity, RegisterInput, LoginInput, JWT claims refactoring
   → contracts/: Domain, use case, and auth interface contracts
   → research.md: Clone-and-extend model decisions and current state analysis
3. Generate tasks by category:
   → Setup: Backup current implementation, validate test environment
   → Tests: Update domain tests, use case tests, JWT tests (TDD approach)
   → Domain: Remove SiteID from User entity and validation
   → Use Cases: Update RegisterInput/LoginInput, remove site_id usage
   → Authentication: Simplify JWT claims to user-only context
   → Repository: Remove site-scoped operations from interfaces and implementations
   → Integration: Verify all components work together without site_id
   → Polish: Run full test suite, validate 80%+ coverage, update documentation
4. Apply task rules:
   → Test files = mark [P] for parallel execution
   → Domain/usecase files = sequential (shared dependencies)
   → JWT package = isolated, can be parallel with domain changes
   → Integration tests run after all implementation complete
5. Number tasks sequentially (T001, T002...)
6. Follow TDD: Update tests first, then implementation
7. Validate Constitutional compliance throughout
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Phase 3.1: Setup & Preparation
- [x] T001 Create backup branch of current multi-tenant implementation
- [x] T002 Validate test environment and dependencies (Go 1.25+, testify, GORM v1.26+)
- [x] T003 [P] Update .gitignore and project documentation for refactoring branch

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be updated and MUST FAIL before ANY implementation changes**

### Domain Tests
- [x] T004 [P] Update User entity tests in `tests/user_domain_test.go` - remove SiteID validation tests
- [x] T005 [P] Update User constructor tests in `tests/user_domain_test.go` - remove siteID parameter
- [x] T006 [P] Update User validation tests in `tests/user_domain_test.go` - remove SiteID required field tests

### Use Case Tests  
- [x] T007 [P] Update RegisterInput validation tests in `internal/usecase/user/register_test.go`
- [x] T008 [P] Update LoginInput validation tests in `internal/usecase/user/login_test.go`
- [x] T009 [P] Update registration use case tests in `internal/usecase/user/register_test.go` - remove site isolation scenarios
- [x] T010 [P] Update login use case tests in `internal/usecase/user/login_test.go` - remove site isolation scenarios

### Authentication Tests
- [x] T011 [P] Update JWT claims tests in `pkg/auth/jwt_test.go` - remove SiteID claim tests
- [x] T012 [P] Update token generation tests in `pkg/auth/jwt_test.go` - remove siteID parameter
- [x] T013 [P] Update token validation tests in `pkg/auth/jwt_test.go` - expect user-only claims

### Integration Tests
- [x] T014 [P] Update user repository interface tests - remove site-scoped operations (mock repositories updated)
- [x] T015 [P] Create new integration tests for clone-and-extend model validation (added to domain tests)

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Domain Layer Updates
- [x] T016 Update User entity in `internal/domain/user.go` - remove SiteID field
- [x] T017 Update NewUser constructor in `internal/domain/user.go` - remove siteID parameter
- [x] T018 Update User.Validate() method in `internal/domain/user.go` - remove SiteID validation
- [x] T019 Update UserRepository interface in `internal/domain/user.go` - remove siteID from FindByEmail

### Use Case Layer Updates
- [x] T020 Update RegisterInput struct in `internal/usecase/user/register.go` - remove SiteID field
- [x] T021 Update RegisterUsecase.Execute() in `internal/usecase/user/register.go` - remove site_id usage
- [x] T022 Update LoginInput struct in `internal/usecase/user/login.go` - remove SiteID field  
- [x] T023 Update LoginUsecase.Execute() in `internal/usecase/user/login.go` - remove site_id usage

### Authentication Layer Updates
- [x] T024 [P] Update JWTClaims struct in `pkg/auth/jwt.go` - remove SiteID field
- [x] T025 [P] Update GenerateToken function in `pkg/auth/jwt.go` - remove siteID parameter
- [x] T026 [P] Update ValidateToken function in `pkg/auth/jwt.go` - handle user-only claims

### Legacy Use Case Updates
- [x] T027 Update UserUseCase.CreateUser() in `internal/usecase/user_usecase.go` - remove siteID parameter
- [x] T028 Update all repository method calls in `internal/usecase/user_usecase.go` - remove site-scoped operations

### Persistence Layer Updates (CRITICAL - Missing Tasks)
- [x] T028a Update GormUser model in `adapters/persistence/user_repository.go` - remove SiteID field
- [x] T028b Update UserRepository.Create() in `adapters/persistence/user_repository.go` - remove SiteID mapping
- [x] T028c Update UserRepository.FindByEmail() in `adapters/persistence/user_repository.go` - remove siteID parameter and site filtering

## Phase 3.4: Integration & Validation
- [x] T029 Run domain tests and verify all pass with updated User entity
- [x] T030 Run use case tests and verify registration/login work without site context
- [x] T031 Run JWT tests and verify tokens contain only user ID claims
- [x] T032 Run integration tests and verify end-to-end flow works
- [x] T033 Validate error messages updated to remove site context references

## Phase 3.5: Polish & Documentation
- [x] T034 [P] Run full test suite and verify 80%+ coverage maintained
- [x] T035 [P] Update domain error messages in `internal/domain/user.go` - remove site references  
- [x] T036 [P] Search codebase for remaining site_id/siteID/SiteID references and clean up
- [x] T037 [P] Update README.md with clone-and-extend model documentation
- [x] T038 Validate Constitutional compliance (hexagonal architecture, TDD, performance standards)
- [x] T039 Create migration guide for existing multi-tenant deployments
- [x] T040 Final integration test with complete user registration and authentication flow

## Dependencies
- Setup (T001-T003) before all other phases
- Tests (T004-T015) before implementation (T016-T028)
- Domain updates (T016-T019) before use case updates (T020-T023, T027-T028)  
- All implementation (T016-T028) before integration (T029-T033)
- Integration (T029-T033) before polish (T034-T040)

## Parallel Execution Examples

### Phase 3.2: Test Updates (Can run simultaneously)
```
Task: "Update User entity tests in tests/user_domain_test.go - remove SiteID validation tests"
Task: "Update RegisterInput validation tests in internal/usecase/user/register_test.go"  
Task: "Update LoginInput validation tests in internal/usecase/user/login_test.go"
Task: "Update JWT claims tests in pkg/auth/jwt_test.go - remove SiteID claim tests"
Task: "Update user repository interface tests - remove site-scoped operations"
```

### Phase 3.3: JWT Updates (Independent of domain changes)
```
Task: "Update JWTClaims struct in pkg/auth/jwt.go - remove SiteID field"
Task: "Update GenerateToken function in pkg/auth/jwt.go - remove siteID parameter"
Task: "Update ValidateToken function in pkg/auth/jwt.go - handle user-only claims"
```

### Phase 3.5: Documentation (Can run simultaneously)
```
Task: "Run full test suite and verify 80%+ coverage maintained"
Task: "Update domain error messages in internal/domain/user.go - remove site references"
Task: "Search codebase for remaining site_id/siteID/SiteID references and clean up"
Task: "Update README.md with clone-and-extend model documentation"
```

## Validation Checklist
✅ All domain tests updated to remove site_id requirements
✅ All use case DTOs simplified (RegisterInput, LoginInput)  
✅ JWT authentication uses user-only claims
✅ Repository interfaces removed site-scoped operations
✅ 80%+ test coverage maintained throughout refactoring
✅ Constitutional compliance verified (hexagonal architecture, TDD, performance)
✅ Clone-and-extend model fully implemented per Constitution v1.2.0