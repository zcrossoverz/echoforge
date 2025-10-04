# Tasks: Configuration and Logging Infrastructure

**Input**: - [x] T013: Context propagation utilities (sequential after T011)esign documents from `C:\Users\Nhan\go\src\echoforge\specs\006-define-config-and\`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Tech stack: Go 1.25+, Viper v1.19.0, Zap v1.27.0, go-playground/validator/v10
   → Structure: Single modular monolith, internal/config/, internal/logging/
2. Load design documents:
   → data-model.md: Config, Logger, ValidationRule, LogEntry entities
   → contracts/: config.md (NewConfig factory), logging.md (NewLogger factory)
   → research.md: Viper integration, Zap setup, security sanitization, hot-reload
3. Generate tasks by category:
   → Setup: Go dependencies, project structure, Viper/Zap/validator integration
   → Tests: TDD with Testify (>80% coverage), contract tests, integration tests  
   → Core: config/logging entities, factory functions, validation rules
   → Integration: Wire DI, Gin middleware, context propagation, security
   → Performance: Benchmarks, hot-reload, memory optimization
   → Polish: Documentation, quickstart validation, backward compatibility
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
   → Security sanitization mandatory
   → >80% test coverage requirement
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → Config contract has tests? ✓
   → Logging contract has tests? ✓
   → All entities have models? ✓
   → Security requirements covered? ✓
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Go modular monolith**: `internal/`, `configs/`, root-level tests
- All paths relative to repository root: `C:\Users\Nhan\go\src\echoforge\`

## Phase 3.1: Setup
- [x] **T001** Update go.mod with required dependencies: Viper v1.19.0, Zap v1.27.0, go-playground/validator/v10, and run `go mod tidy`
- [x] **T002** [P] Create directory structure: `internal/config/`, `internal/logging/`, ensure `configs/` exists
- [x] **T003** [P] Create `configs/config.yaml` template with default values for DB_DSN, JWT_SECRET, LOG_LEVEL, ENABLE_HOT_RELOAD

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] **T004** [P] Contract test for NewConfig() factory in `internal/config/config_test.go` - test env loading, YAML loading, validation, defaults
- [x] **T005** [P] Contract test for NewLogger() factory in `internal/logging/logging_test.go` - test level configuration, JSON/console formats, context propagation
- [x] **T006** [P] Integration test for config hot-reload in `tests/config_integration_test.go` - test file watching, validation on reload, callback execution
- [x] **T007** [P] Integration test for logging security sanitization in `tests/logging_integration_test.go` - test sensitive field filtering, context propagation, performance

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] **T008** [P] Config struct and validation rules in `internal/config/config.go` - implement Config struct with mapstructure and validate tags
- [x] **T009** [P] Logger configuration types in `internal/logging/types.go` - implement LogEntry, LoggerConfig, SensitiveFieldFilter structs
- [x] **T010** NewConfig() factory function in `internal/config/config.go` - implement Viper integration, env/YAML/defaults precedence, validation
- [x] **T011** NewLogger() factory function in `internal/logging/logging.go` - implement Zap configuration, JSON/console formats, level setting
- [x] **T012** Security sanitization in `internal/logging/sanitizer.go` - implement sensitive field detection and [REDACTED] replacement
- [x] **T013** Context propagation utilities in `internal/logging/context.go` - implement request ID injection, context-aware logging methods
- [x] **T014** Hot-reload functionality in `internal/config/watcher.go` - implement file watching with debouncing, validation on reload

## Phase 3.4: Integration
- [x] **T015** Wire dependency injection providers in `internal/di/providers.go` - create provider set for Config and Logger with Wire
- [x] **T016** Gin logging middleware in `adapters/http/middleware/logging.go` - implement request ID generation, request/response logging
- [x] **T017** Context utilities for existing handlers in `pkg/common/context.go` - implement request ID extraction, logger injection helpers
- [x] **T018** Update existing authentication to use new config in `pkg/auth/jwt.go` - use Config.JWTSecret from new config system
- [x] **T019** Backward compatibility layer in `internal/config/legacy.go` - ensure existing config access patterns continue working

## Phase 3.5: Performance & Validation
- [x] **T020** [P] Performance benchmarks in `internal/config/config_benchmark_test.go` - test config loading <5s, hot-reload <1s, memory usage
- [ ] **T021** [P] Performance benchmarks in `internal/logging/logging_benchmark_test.go` - test 1000+ logs/sec, memory footprint <50MB, context overhead
- [ ] **T022** [P] Security audit tests in `tests/security_test.go` - verify no sensitive data in logs, test all sanitization patterns

## Phase 3.6: Polish
- [ ] **T023** [P] Unit tests for edge cases in `internal/config/edge_cases_test.go` - missing files, malformed YAML, invalid env values
- [ ] **T024** [P] Unit tests for error handling in `internal/logging/error_test.go` - invalid log levels, context propagation failures, sanitization edge cases
- [ ] **T025** Execute quickstart validation scenarios from `specs/006-define-config-and/quickstart.md` - run all examples, verify outputs
- [ ] **T026** Update application main.go to use new config/logging system - replace existing config loading with new factories
- [ ] **T027** [P] Add godoc comments to all public interfaces and ensure go vet passes
- [ ] **T028** Verify >80% test coverage with `go test -cover` and generate coverage reports

## Dependencies
- Setup (T001-T003) before everything
- Tests (T004-T007) before implementation (T008-T014)
- T008 blocks T010 (Config struct needed for factory)
- T009 blocks T011, T012, T013 (types needed for logger implementation)
- T010 blocks T014 (Config loading needed for hot-reload)
- T011 blocks T012, T013 (Logger needed for sanitization and context)
- Core implementation (T008-T014) before integration (T015-T019)
- Integration (T015-T019) before performance/polish (T020-T028)

## Parallel Execution Examples

### Phase 3.2 - TDD Test Creation (All Parallel)
```bash
# Launch T004-T007 together (different files):
Task: "Contract test for NewConfig() factory in internal/config/config_test.go"
Task: "Contract test for NewLogger() factory in internal/logging/logging_test.go"  
Task: "Integration test for config hot-reload in tests/config_integration_test.go"
Task: "Integration test for logging security in tests/logging_integration_test.go"
```

### Phase 3.3 - Core Implementation (T008-T009 Parallel, then rest sequentially)
```bash
# First wave (T008-T009):
Task: "Config struct and validation rules in internal/config/config.go"
Task: "Logger configuration types in internal/logging/types.go"

# Wait for completion, then sequential implementation (T010-T014)
```

### Phase 3.6 - Polish (T023-T024, T027 Parallel)
```bash  
# Launch independent polish tasks:
Task: "Unit tests for edge cases in internal/config/edge_cases_test.go"
Task: "Unit tests for error handling in internal/logging/error_test.go"
Task: "Add godoc comments to all public interfaces"
```

## Testing Requirements
- **Minimum Coverage**: >80% as specified in constitutional requirements
- **TDD Approach**: All tests must be written first and must fail before implementation
- **Test Categories**: Contract tests, integration tests, performance benchmarks, security audits, edge case handling
- **Performance Targets**: Config loading <5s, logging 1000+ entries/sec, hot-reload <1s, memory <50MB

## Security Requirements  
- **OWASP Compliance**: All sensitive data sanitization mandatory
- **Sensitive Fields**: password, secret, token, dsn, key, auth, credential (case-insensitive)
- **Sanitization**: Replace sensitive values with [REDACTED] in all log output
- **Validation**: All config inputs validated before use, clear error messages for failures

## File Impact Summary
**New Files Created** (22 files):
- `internal/config/config.go`, `internal/config/config_test.go`, `internal/config/watcher.go`, `internal/config/legacy.go`, `internal/config/edge_cases_test.go`, `internal/config/config_benchmark_test.go`
- `internal/logging/logging.go`, `internal/logging/logging_test.go`, `internal/logging/types.go`, `internal/logging/sanitizer.go`, `internal/logging/context.go`, `internal/logging/error_test.go`, `internal/logging/logging_benchmark_test.go`
- `tests/config_integration_test.go`, `tests/logging_integration_test.go`, `tests/security_test.go`
- `adapters/http/middleware/logging.go`
- `internal/di/providers.go`
- `pkg/common/context.go`
- `configs/config.yaml`

**Modified Files** (2 files):
- `go.mod` (add dependencies)
- `pkg/auth/jwt.go` (use new config)
- `cmd/server/main.go` (integrate new config/logging)

## Validation Checklist
*GATE: Checked before task execution*

- [x] All contracts have corresponding tests (T004-T005)
- [x] All entities have model tasks (T008-T009)  
- [x] All tests come before implementation (T004-T007 before T008-T014)
- [x] Parallel tasks truly independent (different files, no shared dependencies)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] Security requirements explicit in multiple tasks
- [x] Performance requirements testable with benchmarks
- [x] >80% coverage requirement specified and testable

---

**Task Generation Status**: ✅ Complete - 28 tasks generated with clear dependencies and parallel execution guidance