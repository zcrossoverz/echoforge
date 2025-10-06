
# Implementation Plan: Configuration and Logging Infrastructure

**Branch**: `006-define-config-and` | **Date**: October 4, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `C:\Users\Nhan\go\src\echoforge\specs\006-define-config-and\spec.md`

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
Configuration management system that loads settings from environment variables and YAML files with validation and defaults, plus structured JSON logging with configurable levels and context propagation. Focus on OWASP-compliant security (no secrets in logs), hot-reload capabilities for development, and dependency injection support via Wire for the modular monolith architecture.

## Technical Context
**Language/Version**: Go 1.25+  
**Primary Dependencies**: Viper v1.19.0 (config), Zap v1.27.0 (logging), go-playground/validator/v10 (validation)  
**Storage**: Configuration files (YAML) and environment variables, no database storage for this feature  
**Testing**: Testify with TDD approach, >80% coverage requirement  
**Target Platform**: Cross-platform (Linux/Windows/macOS) server environments
**Project Type**: Single modular monolith - extends existing echoforge core  
**Performance Goals**: 1000+ log entries/second, config loading <5 seconds, hot-reload <1 second  
**Constraints**: Binary size <20MB total, memory footprint <50MB, lean MVP ~100 LOC  
**Scale/Scope**: Configuration for multi-site deployment, structured logging for production monitoring

**User Arguments Integration**: Config setup in internal/config/config.go with Viper v1.19.0 for env/yaml loading (DB_DSN, JWT_SECRET, LOG_LEVEL with defaults). Logging in internal/logging/logging.go with Zap v1.27.0 for structured JSON output. Factories: NewConfig() and NewLogger(). Context-aware logging with request ID propagation. Wire DI integration. OWASP compliance with no sensitive data logging. TDD with config_test.go and logging_test.go achieving >80% coverage.

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **Modular monolith with hexagonal architecture**: Config/logging are infrastructure concerns, properly placed in internal/ packages  
✅ **GORM v1.26+ with golang-migrate**: N/A - this feature handles configuration loading, no database persistence required  
✅ **Gin v1.10+ with versioned APIs**: N/A - this feature provides infrastructure services to the existing Gin API layer  
✅ **TDD with Testify required, 80%+ test coverage**: Explicitly required in spec (config_test.go, logging_test.go >80% coverage)  
✅ **Clone-and-extend model**: Config supports per-site customization via config.yaml/env without core modification  
✅ **Auth compatibility**: Loads JWT_SECRET for existing bcrypt+JWT auth system  
✅ **Reusable core design**: Factory pattern (NewConfig, NewLogger) supports dependency injection via Wire  
✅ **SemVer compliance**: Backward-compatible interfaces, no breaking changes to existing config/logging  
✅ **Performance**: Meets 1000+ log entries/second requirement, config loading <5 seconds  
✅ **Security**: OWASP compliance explicit - sanitizes sensitive data (DB_DSN, JWT_SECRET) from logs  
✅ **Lean MVP**: ~100 LOC target, focused scope on essential config/logging infrastructure  
✅ **Dependencies**: Viper v1.19.0, Zap v1.27.0, validator/v10 pinned in go.mod  
✅ **Zero-downtime**: Hot-reload support for development, no schema changes required

**PASS** - All constitutional requirements satisfied or N/A for infrastructure feature scope.

### Post-Design Re-evaluation
✅ **Design artifacts completed**: research.md, data-model.md, contracts/, quickstart.md, agent context updated  
✅ **Factory patterns implemented**: NewConfig(), NewLogger() for dependency injection via Wire  
✅ **Security requirements met**: Sensitive data sanitization, OWASP compliance in logging design  
✅ **Performance targets achievable**: 1000+ logs/sec with Zap, <5s config load with Viper caching  
✅ **TDD approach defined**: Comprehensive test scenarios in contracts and quickstart for >80% coverage  
✅ **Integration preserves boundaries**: Clean hexagonal architecture with infrastructure in internal/ packages

**FINAL PASS** - Post-design constitutional compliance confirmed.

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
├── config/
│   ├── config.go           # Config struct, NewConfig() factory, Viper integration
│   └── config_test.go      # Config validation tests (env, yaml, defaults)
├── logging/
│   ├── logging.go          # Logger setup, NewLogger() factory, Zap integration  
│   └── logging_test.go     # Logging format and level tests
└── domain/                 # Existing domain entities (unchanged)

configs/
└── config.yaml             # Default configuration template

tests/                      # Integration tests for config+logging interaction
├── config_integration_test.go
└── logging_integration_test.go

go.mod                      # Updated with Viper v1.19.0, Zap v1.27.0, validator/v10
```

**Structure Decision**: Single Go modular monolith extending existing echoforge structure. Configuration and logging are infrastructure concerns placed in `internal/config/` and `internal/logging/` respectively, following hexagonal architecture principles. This maintains clean separation from domain logic while providing shared infrastructure services to all application layers.

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
- Load `.specify/templates/tasks-template.md` as base
- Generate TDD tasks from contracts: config.md and logging.md interfaces
- Create test files first: config_test.go, logging_test.go (>80% coverage requirement)
- Implementation tasks: config.go, logging.go with factory functions
- Integration tasks: Wire dependency injection, Gin middleware, GORM integration
- Validation tasks: Quickstart scenarios, performance benchmarks, security tests

**Ordering Strategy**:
- **Phase A (Tests First)**: Contract test creation [P], validation test setup [P]
- **Phase B (Core Implementation)**: Config factory → Logger factory → Security sanitization
- **Phase C (Integration)**: Wire DI setup → Gin middleware → Context propagation
- **Phase D (Validation)**: Quickstart execution → Performance testing → Security audit
- Mark [P] for parallel execution (independent files)

**Specific Task Categories**:
1. **Test Infrastructure** (6 tasks): Contract tests, validation tests, integration tests
2. **Configuration System** (4 tasks): Viper integration, validation, hot-reload, defaults
3. **Logging System** (5 tasks): Zap setup, sanitization, context propagation, levels
4. **Integration Layer** (4 tasks): Wire DI, Gin middleware, GORM compatibility, backwards compatibility
5. **Validation & Performance** (3 tasks): Quickstart validation, performance benchmarks, security audit

**Estimated Output**: 22 numbered, ordered tasks with clear TDD progression in tasks.md

**Dependencies Identified**:
- go.mod updates for Viper v1.19.0, Zap v1.27.0, validator/v10
- configs/config.yaml template creation  
- Wire provider set updates for DI integration
- Existing JWT/auth system integration points

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
- [x] Complexity deviations documented (N/A - no deviations)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
