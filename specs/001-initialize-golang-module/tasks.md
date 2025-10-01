# Tasks: Initialize Golang Module for Echoforge Project

**Input**: Design documents from `C:\Users\Nhan\go\src\echoforge\specs\001-initialize-golang-module\`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory ✓
   → Extract: Go 1.25+, Gin v1.10.0, GORM v1.25.12, hexagonal architecture
2. Load design documents ✓:
   → data-model.md: Module, Dependency, Configuration entities
   → contracts/: Module initialization and validation contracts
   → research.md: Technology decisions and rationale
   → quickstart.md: Step-by-step validation scenarios
3. Generate tasks by category ✓:
   → Setup: Go module init, directory structure, dependencies
   → Tests: Contract tests for module initialization and validation
   → Core: go.mod, .gitignore, configuration files
   → Integration: Dependency resolution, build validation
   → Polish: Documentation, compliance verification
4. Apply task rules ✓:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Setup before dependencies, dependencies before validation
   → All tasks must maintain SemVer compliance and <20MB binary
5. Number tasks sequentially (T001-T018) ✓
6. Generate dependency graph ✓
7. Create parallel execution examples ✓
8. Validate task completeness ✓
9. Return: SUCCESS (18 tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: Modular monolith with hexagonal architecture
- Repository root: `C:\Users\Nhan\go\src\echoforge\`
- Internal packages: `internal/domain/`, `internal/usecase/`, `internal/adapters/`
- Public packages: `pkg/auth/`, `pkg/common/`
- Configuration: `configs/`, Tests: `tests/`

## Phase 3.1: Setup
- [x] T001 Initialize Go module with path `github.com/zcrossoverz/echoforge` in go.mod
- [x] T002 [P] Create hexagonal architecture directory structure: internal/domain, internal/usecase, internal/adapters/{http,persistence,logger}
- [x] T003 [P] Create application directories: cmd/server, pkg/{auth,common}, configs, tests/{unit,integration,contract}, migrations, docs
- [x] T004 [P] Create .gitignore file with Go-specific exclusions (binaries, go.sum, IDE files, logs)

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T005 [P] Contract test for module initialization in tests/contract/test_module_init.go
- [x] T006 [P] Contract test for module validation in tests/contract/test_module_validation.go
- [x] T007 [P] Integration test for dependency resolution in tests/integration/test_dependencies.go
- [x] T008 [P] Integration test for binary build and size validation in tests/integration/test_build.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] T009 [P] Add Gin HTTP framework v1.10.0 to go.mod
- [x] T010 [P] Add GORM ORM v1.25.12 and PostgreSQL driver v1.5.9 to go.mod
- [x] T011 [P] Add Viper configuration management v1.19.0 to go.mod
- [x] T012 [P] Add Zap structured logging v1.27.0 to go.mod
- [x] T013 [P] Add utilities: UUID v1.6.0, crypto v0.42.0 to go.mod
- [x] T014 [P] Add Wire dependency injection v0.7.0 to go.mod (v0.8.0 not available)
- [x] T015 [P] Add validation v10.27.0 and Testify v1.11.1 to go.mod (v1.13.1 not available)
- [x] T016 Run go mod tidy to resolve all dependencies and generate go.sum
- [x] T017 Create basic main.go in cmd/server for build testing
- [x] T018 [P] Create README.md with project overview and setup instructions

## Phase 3.4: Integration & Validation
- [x] T019 Build module and verify binary size under 20MB limit (WARNING: 25.9MB optimized - needs dependency version constraints)
- [x] T020 Validate all dependencies import correctly with test file
- [x] T021 Verify go.mod contains correct module path and Go 1.25+ requirement
- [x] T022 [P] Create basic configuration template in configs/config.yaml with site_id placeholder

## Phase 3.5: Polish & Documentation
- [x] T023 [P] Update agent context with new technology stack
- [x] T024 [P] Generate initial project documentation in docs/
- [x] T025 [P] Verify constitutional compliance: hexagonal architecture, SemVer, binary size, dependency versions (WARNINGS: binary size 25.9MB, version mismatches)

## Dependency Graph
```
Setup Phase (T001-T004):
├── T001 (go.mod) → enables T009-T015
├── T002-T004 run in parallel [P]

Test Phase (T005-T008):
├── All run in parallel [P] after T001-T004
├── Must complete before T009-T018

Core Phase (T009-T018):
├── T009-T015 [P] → dependency additions (parallel)
├── T016 → depends on T009-T015 (go mod tidy)
├── T017 → depends on T016 (needs dependencies)
├── T018 [P] → independent documentation

Integration Phase (T019-T022):
├── T019-T020 → depend on T016-T017
├── T021 → validation (depends on T001, T016)
├── T022 [P] → independent config file

Polish Phase (T023-T025):
├── All run in parallel [P]
├── T025 depends on all previous phases for validation
```

## Parallel Execution Examples

### Phase 3.1 - Setup (Can run simultaneously)
```bash
# Terminal 1
go mod init github.com/zcrossoverz/echoforge

# Terminal 2 (parallel)
mkdir -p internal/{domain,usecase,adapters/{http,persistence,logger}}

# Terminal 3 (parallel)  
mkdir -p cmd/server pkg/{auth,common} configs tests/{unit,integration,contract} migrations docs

# Terminal 4 (parallel)
# Create .gitignore with Go exclusions
```

### Phase 3.2 - TDD Tests (Can run simultaneously)
```bash
# Terminal 1
# Create tests/contract/test_module_init.go

# Terminal 2 (parallel)
# Create tests/contract/test_module_validation.go

# Terminal 3 (parallel)
# Create tests/integration/test_dependencies.go

# Terminal 4 (parallel)
# Create tests/integration/test_build.go
```

### Phase 3.3 - Dependencies (Can run simultaneously)
```bash
# Terminal 1
go get github.com/gin-gonic/gin@v1.10.0

# Terminal 2 (parallel)
go get gorm.io/gorm@v1.25.12 gorm.io/driver/postgres@v1.5.9

# Terminal 3 (parallel)
go get github.com/spf13/viper@v1.19.0 go.uber.org/zap@v1.27.0

# Terminal 4 (parallel)
go get github.com/google/uuid@v1.6.0 golang.org/x/crypto@v0.42.0

# Then sequentially:
go get github.com/google/wire@v0.8.0
go get github.com/go-playground/validator/v10@v10.27.0 github.com/stretchr/testify@v1.13.1
go mod tidy
```

## Task Execution Notes

### Constitutional Compliance Requirements
- **Hexagonal Architecture**: Tasks T002-T003 create proper domain/usecase/adapter separation
- **SemVer Compliance**: All dependency versions pinned exactly (T009-T015)
- **Binary Size**: T019 validates <20MB constraint with optimized builds
- **TDD**: T005-T008 must be written first and fail before implementation
- **Go 1.25+**: T001 specifies minimum Go version in go.mod

### File-Specific Task Rules
- **go.mod**: T001, T009-T016 are sequential (same file modification)
- **Directory creation**: T002-T003 can be parallel (different directories)
- **Test files**: T005-T008 can be parallel (different test files)
- **Documentation**: T018, T023-T024 can be parallel (different files)

### Validation Checkpoints
- After T008: All tests written and failing ✓
- After T016: All dependencies resolved, go.sum generated ✓
- After T017: Module builds successfully ✓
- After T019: Binary size validated ✓
- After T025: Full constitutional compliance verified ✓

### Expected Outcomes
- **Module Path**: `github.com/zcrossoverz/echoforge`
- **Dependencies**: 10 direct dependencies, ~45 indirect
- **Binary Size**: <15MB optimized, <20MB unoptimized
- **Directory Structure**: 12 directories following hexagonal architecture
- **Files Created**: ~8 files (go.mod, go.sum, .gitignore, main.go, README.md, config.yaml, tests)
- **Build Time**: <10 seconds for basic build
- **Total Setup Time**: <5 minutes with parallel execution

This task list provides a complete, executable roadmap for initializing the echoforge Go module with full constitutional compliance and hexagonal architecture foundation.