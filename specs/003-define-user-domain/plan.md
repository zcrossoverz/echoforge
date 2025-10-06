
# Implementation Plan: User Domain Entity and Repository

**Branch**: `003-define-user-domain` | **Date**: 2025-10-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-define-user-domain/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Implement a foundational User domain entity and repository for multi-tenant authentication system. The User entity will include unique identifier, site isolation, email validation, secure password storage, and timestamps. Repository interface will support creating users, retrieving by email within site context, and handling validation errors with complete tenant isolation.

## Technical Context
**Language/Version**: Go 1.25+  
**Primary Dependencies**: GORM v1.26+ (ORM), Testify (testing), UUID package (unique identifiers), bcrypt (password hashing)  
**Storage**: PostgreSQL 16+ with GORM ORM, golang-migrate for migrations  
**Testing**: Testify with TDD approach, 80%+ coverage requirement  
**Target Platform**: Linux server, Docker containers with rolling updates
**Project Type**: single (modular monolith with hexagonal architecture)  
**Performance Goals**: 1000+ concurrent users per site, optimized multi-tenant queries  
**Constraints**: Multi-site tenant isolation via site_id, backward compatibility (SemVer), OWASP Top 10 security  
**Scale/Scope**: Foundation for multi-tenant user management across 10+ sites, lean MVP approach (500-1000 LOC)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Initial Check (Pre-Phase 0)**:
- ✅ Modular monolith with hexagonal architecture (User entity in internal/domain, repository interface)
- ✅ GORM v1.26+ with golang-migrate (User persistence layer using GORM)
- ⚠️ Gin v1.10+ with versioned APIs (Not applicable - this is domain layer only, no HTTP endpoints)
- ✅ TDD with Testify required, 80%+ test coverage mandatory
- ✅ Multi-site tenant isolation via `site_id` in config and DB queries
- ✅ Auth: bcrypt + JWT with unique email constraint (User entity includes password hash validation)
- ✅ Reusable core design: config override without core modification
- ✅ SemVer compliance: backward compatibility for interfaces
- ✅ Performance: 1000+ concurrent users/site capability (optimized queries with site_id)
- ✅ Security: OWASP Top 10 compliance, input validation (email/password validation)
- ✅ Lean MVP: 500-1000 LOC limit, YAGNI principles (focused on core User entity only)
- ✅ All dependencies documented in go.mod with version constraints
- ⚠️ Zero-downtime deployments via Docker (Not applicable - domain layer implementation)

**Post-Design Check (After Phase 1)**:
- ✅ Clean architecture boundaries maintained (domain/usecase/adapters separation)
- ✅ Repository interface follows contract specification with proper error handling
- ✅ Database schema includes proper constraints and indexes for performance
- ✅ Validation rules implement constitutional security requirements
- ✅ Multi-tenant isolation enforced at entity and repository level
- ✅ Test-driven design with comprehensive unit and integration test scenarios
- ✅ Error handling distinguishes between domain and infrastructure concerns

**Final Status**: PASS - All constitutional requirements satisfied for domain layer implementation

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
internal/
├── domain/
│   └── user.go              # User entity with validation (pure domain)
└── usecase/
    └── user_usecase.go      # Business logic with repository interface

adapters/
├── persistence/
│   └── user_repository.go   # GORM implementation of repository interface
└── http/
    └── (future API handlers)

tests/
├── user_domain_test.go      # Unit tests for User entity
├── user_usecase_test.go     # Unit tests for use cases
└── user_repository_test.go  # Integration tests for repository

migrations/
└── 002_create_users_table.up.sql  # Database migration for users table
└── 002_create_users_table.down.sql

pkg/
└── (shared utilities as needed)
```

**Structure Decision**: Go modular monolith with hexagonal architecture. User entity lives in pure domain layer (`internal/domain`), business logic in use cases (`internal/usecase`), and GORM persistence adapter in `adapters/persistence`. This maintains clean boundaries and prepares for future HTTP API layer.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType copilot`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base template
- Generate tasks from Phase 1 design artifacts:
  - `data-model.md` → User entity implementation tasks
  - `contracts/user_repository.md` → Repository interface and mock tasks
  - `quickstart.md` → Validation and integration test tasks
- Domain layer tasks (TDD approach):
  - Contract tests for repository interface [P]
  - User entity implementation with validation [P]  
  - Unit tests for User entity [P]
  - Mock repository implementation [P]
- Persistence layer tasks:
  - Database migration files 
  - GORM repository implementation
  - Integration tests with real database
- Use case layer tasks:
  - Business logic implementation using repository interface
  - Use case unit tests with mock repository

**Ordering Strategy**:
- **Phase 1 (Domain Foundation)**: Pure domain layer (no dependencies)
  1. Repository interface definition
  2. User entity with validation
  3. Domain unit tests
  4. Mock repository implementation
- **Phase 2 (Persistence Layer)**: Database integration
  5. Database migration files
  6. GORM repository implementation  
  7. Repository integration tests
- **Phase 3 (Business Logic)**: Use case implementation
  8. User creation use case
  9. User retrieval use case
  10. Use case unit tests
- **Phase 4 (Validation)**: End-to-end validation
  11. Contract test verification
  12. Quickstart guide validation
  13. Performance benchmarks

**Parallelization Strategy**:
- Mark [P] for independent file creation tasks
- Sequential execution for dependent layers (domain → persistence → usecase)
- Parallel test creation within each layer

**Estimated Task Breakdown**:
- Domain layer: 6-8 tasks (entity, interface, tests, mocks)
- Persistence layer: 4-6 tasks (migrations, repository, integration tests)
- Use case layer: 4-6 tasks (business logic, unit tests)
- Validation: 3-4 tasks (contract verification, quickstart validation)
- **Total**: 17-24 numbered, ordered tasks in tasks.md

**Quality Gates**:
- Each implementation task must have corresponding test task
- TDD enforcement: failing tests before implementation
- 80%+ coverage verification task included
- Constitutional compliance validation task included

**File Structure Targets**:
```
internal/domain/user.go                    # User entity + repository interface
adapters/persistence/user_repository.go   # GORM implementation
tests/user_domain_test.go                 # Domain unit tests
tests/user_repository_test.go             # Repository integration tests
tests/user_usecase_test.go                # Use case unit tests
migrations/002_create_users_table.*.sql   # Database schema
```

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - research.md generated
- [x] Phase 1: Design complete (/plan command) - data-model.md, contracts/, quickstart.md created
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS  
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none required)

**Artifacts Generated**:
- [x] `research.md`: Technical decisions and dependency analysis
- [x] `data-model.md`: User entity specification and database schema
- [x] `contracts/user_repository.md`: Repository interface contract with mock implementation
- [x] `quickstart.md`: Complete implementation guide with examples
- [x] Agent context updated: `.github/copilot-instructions.md`

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
