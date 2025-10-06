
# Implementation Plan: Register and Login Authentication Usecases

**Branch**: `004-define-register-and` | **Date**: October 2, 2025 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-define-register-and/spec.md`

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
Implement Register and Login usecases for echoforge's multi-tenant authentication system. Users can create accounts and authenticate on any site while maintaining complete data isolation via `site_id`. System provides secure OWASP-compliant authentication using bcrypt password hashing and JWT tokens with 24-hour expiration. Architecture follows hexagonal pattern with injectable usecases, GORM persistence adapters, and comprehensive TDD test coverage >80%.

## Technical Context
**Language/Version**: Go 1.25+  
**Primary Dependencies**: gin v1.10.0, gorm.io/gorm v1.26.12, go-playground/validator/v10 v10.27.0, golang.org/x/crypto v0.42.0, github.com/golang-jwt/jwt/v5, testify v1.13.1  
**Storage**: PostgreSQL 16+ with GORM ORM, existing user domain/repository from Task 1.2  
**Testing**: Testify with TDD approach, 80%+ coverage requirement  
**Target Platform**: Linux server (Docker containers), multi-site deployment
**Project Type**: Single backend project (modular monolith with hexagonal architecture)  
**Performance Goals**: 1000+ concurrent authentication requests per site, <2 seconds per auth operation  
**Constraints**: OWASP Top 10 compliance, bcrypt cost ≥12, JWT 24h expiration, multi-tenant isolation  
**Scale/Scope**: Lean MVP ~150 LOC, 2 new usecases (Register, Login), JWT secret via env var

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ Modular monolith with hexagonal architecture: Usecases in `internal/usecase/user/`, leveraging existing domain in `internal/domain/`
- ✅ GORM v1.26+ with golang-migrate: Uses existing user repository and domain, no new migrations needed
- ⚠️ Gin v1.10+ with versioned APIs: Not applicable for this phase (usecases only, HTTP layer deferred)
- ✅ TDD with Testify required, 80%+ test coverage: Comprehensive test suite planned for both usecases
- ✅ Multi-site tenant isolation via `site_id`: All operations scoped by siteID parameter in usecase inputs
- ✅ Auth: bcrypt + JWT with unique email constraint: bcrypt cost=12, JWT HS256 24h, leverages existing unique email
- ✅ Reusable core design: JWT secret configurable via environment, no core modifications needed
- ✅ SemVer compliance: Builds on existing user domain, backward compatible addition
- ✅ Performance: Leverages existing 1000+ concurrent capability, stateless JWT design
- ✅ Security: Input validation with go-playground/validator/v10, structured error handling
- ✅ Lean MVP: ~150 LOC target, focused on 2 core usecases only
- ✅ Dependencies documented: Existing deps sufficient, golang-jwt/jwt/v5 to be added
- ✅ Zero-downtime deployments: Stateless usecases compatible with rolling updates

**Compliance Status**: PASS - All constitutional requirements satisfied

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
echoforge/
├── internal/
│   ├── domain/
│   │   └── user.go              # Existing user entity from Task 1.2
│   └── usecase/
│       └── user/                # NEW: Authentication usecases
│           ├── register.go      # RegisterUsecase implementation
│           ├── login.go         # LoginUsecase implementation
│           ├── register_test.go # RegisterUsecase TDD tests
│           └── login_test.go    # LoginUsecase TDD tests
├── adapters/
│   └── persistence/
│       └── user_repository.go   # Existing GORM user repository from Task 1.2
├── pkg/
│   ├── auth/
│   │   ├── jwt.go              # Existing JWT utilities
│   │   └── jwt_test.go         # JWT utilities tests
│   └── common/
│       └── logger.go           # Existing logger utilities
├── configs/
│   └── config.yaml             # Existing config with JWT_SECRET addition
├── tests/                      # Existing test infrastructure
├── go.mod                      # Existing dependencies + JWT library
└── go.sum
```

**Structure Decision**: Extending existing hexagonal architecture with new usecase layer. Places authentication business logic in `internal/usecase/user/` while leveraging existing domain entities and repository adapters. Maintains clean separation between business logic (usecases) and external concerns (JWT utilities, persistence).

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
- Load `.specify/templates/tasks-template.md` as base structure
- Generate tasks from Phase 1 design artifacts:
  - `data-model.md` → Input/Output DTO creation tasks [P]
  - `contracts/usecase-contracts.md` → Interface and test creation tasks [P]
  - `quickstart.md` → Validation and integration test tasks
  - `research.md` → JWT utility and dependency integration tasks

**Specific Task Categories**:
1. **Foundation Tasks** (Sequential):
   - Add JWT dependency to go.mod
   - Create JWT utility functions in pkg/auth/jwt.go
   - Add JWT secret configuration to config handling

2. **TDD Test Tasks** (Parallel after foundation):
   - RegisterUsecase interface and failing tests [P]
   - LoginUsecase interface and failing tests [P] 
   - Input validation test suites [P]
   - JWT token generation/validation tests [P]

3. **Implementation Tasks** (Sequential after tests):
   - RegisterUsecase implementation (make tests pass)
   - LoginUsecase implementation (make tests pass)
   - Error handling and validation logic
   - Integration test implementation

4. **Validation Tasks** (Final):
   - Security test suite (multi-tenant isolation)
   - Performance benchmarking
   - End-to-end authentication flow testing

**Ordering Strategy**:
- **TDD Enforcement**: All test tasks before corresponding implementation
- **Dependency Resolution**: JWT utilities before usecases, interfaces before implementations
- **Parallel Opportunities**: Mark independent tasks with [P] for concurrent execution
- **Integration Last**: Full system tests after all units complete

**Estimated Breakdown**:
- Foundation setup: 3-4 tasks
- TDD test creation: 6-8 tasks (parallelizable)
- Implementation: 4-6 tasks
- Validation & integration: 4-5 tasks
- **Total**: 17-23 numbered, dependency-ordered tasks

**Quality Gates**:
- Each implementation task requires corresponding tests to pass
- 80% coverage requirement enforced at task level
- Multi-tenant isolation validated in dedicated security tasks
- Performance benchmarks included for authentication operations

**IMPORTANT**: This phase will be executed by the `/tasks` command, NOT by `/plan`

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
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
