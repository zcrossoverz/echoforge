
# Implementation Plan: Abstract Post System

**Branch**: `007-abstract-post-feature` | **Date**: October 5, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-abstract-post-feature/spec.md`

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
Extensible post management system enabling site creators to build specialized content platforms (blog, manga, news) by extending a flexible base post entity. System supports custom post types, multimedia attachments, scheduling, approval workflows, and bulk operations while maintaining multi-tenancy through the clone-and-extend model.

## Technical Context
**Language/Version**: Go 1.25+ (constitutional requirement)  
**Primary Dependencies**: GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, Testify v1.11+  
**Storage**: PostgreSQL 16+ with separate database per site instance  
**Testing**: Testify framework with TDD methodology, 80%+ coverage requirement  
**Target Platform**: Linux server with Docker deployment, zero-downtime rolling updates
**Project Type**: Single modular monolith with hexagonal architecture  
**Performance Goals**: 1000+ concurrent users per site, 500ms max response time  
**Constraints**: OWASP Top 10 compliance, multi-site isolation, SemVer backward compatibility  
**Scale/Scope**: Extensible post system supporting blog/manga/news use cases, 500-1000 LOC MVP

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ Modular monolith with hexagonal architecture (domain in internal/domain, adapters separated)
- ✅ GORM v1.26+ with golang-migrate required for all Postgres persistence
- ✅ Gin v1.10+ with versioned APIs (/api/v1/) required for all HTTP endpoints
- ✅ TDD with Testify required, 80%+ test coverage mandatory
- ✅ Multi-site tenant isolation via clone-and-extend model (separate DBs per site)
- ✅ Auth: existing bcrypt + JWT system will be leveraged for post author identification
- ✅ Reusable core design: post system extensible via config and interfaces
- ✅ SemVer compliance: backward compatibility for API/config changes
- ✅ Performance: 1000+ concurrent users/site capability (NFR-001 requirement)
- ✅ Security: OWASP Top 10 compliance, input validation for post content
- ✅ Lean MVP: 500-1000 LOC limit, YAGNI principles (focus on core post operations)
- ✅ All dependencies documented in go.mod with version constraints
- ✅ Zero-downtime deployments via Docker rolling updates

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
internal/domain/
├── post.go              # Post entity with validation
├── post_type.go         # PostType entity  
├── post_category.go     # PostCategory entity
├── post_tag.go          # PostTag entity
├── post_attachment.go   # PostAttachment entity
├── post_version.go      # PostVersion entity
├── post_metadata.go     # PostMetadata entity
└── post_repository.go   # Repository interfaces

internal/usecase/
├── post_usecase.go      # Post business logic
├── post_type_usecase.go # PostType management
└── post_search_usecase.go # Search and filtering

adapters/http/
├── handlers/
│   ├── post_handler.go  # Post CRUD endpoints
│   ├── post_type_handler.go # PostType management
│   └── post_search_handler.go # Search endpoints
└── middleware/
    └── post_validation_middleware.go

adapters/persistence/
├── post_repository.go   # GORM implementation
├── post_type_repository.go
└── post_search_repository.go

tests/unit/
├── domain/
├── usecase/
└── handlers/

tests/integration/
└── post_flow_test.go

migrations/
└── 00X_create_post_tables.up.sql
```

**Structure Decision**: Single modular monolith following hexagonal architecture with domain entities in internal/domain, business logic in internal/usecase, and adapters for HTTP/persistence following existing project conventions.

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
- Load `.specify/templates/tasks-template.md` as base framework
- Generate domain entity tasks from data-model.md (7 entities × 2 tasks each = 14 tasks)
- Generate repository interface tasks from data model relationships (7 interfaces = 7 tasks)
- Generate use case tasks from functional requirements FR-001 to FR-015 (15 tasks)
- Generate HTTP handler tasks from API contracts (8 endpoint groups = 8 tasks)
- Generate GORM adapter tasks for persistence (7 repositories = 7 tasks)
- Generate migration tasks from database schema (1 comprehensive migration = 2 tasks)
- Generate contract test tasks from OpenAPI specification (8 endpoint groups = 8 tasks)
- Generate integration test tasks from quickstart scenarios (5 scenarios = 5 tasks)
- Generate performance test tasks from NFR requirements (3 tasks)

**Ordering Strategy**:
- **Phase 1**: Domain entities with validation (tasks 1-14) [P]
- **Phase 2**: Repository interfaces (tasks 15-21) [P] 
- **Phase 3**: Database migrations (tasks 22-23) [sequential dependency]
- **Phase 4**: Use case business logic (tasks 24-38) [depends on domain + repos]
- **Phase 5**: HTTP handlers (tasks 39-46) [depends on use cases]
- **Phase 6**: GORM persistence adapters (tasks 47-53) [P, depends on migrations]
- **Phase 7**: Contract tests (tasks 54-61) [P, depends on handlers]
- **Phase 8**: Integration tests (tasks 62-66) [depends on full stack]
- **Phase 9**: Performance tests (tasks 67-69) [depends on integration tests]

**Parallel Execution Markers**:
- [P] for domain entities (independent files)
- [P] for repository interfaces (independent contracts)
- [P] for GORM adapters (independent implementations)  
- [P] for contract tests (independent endpoint testing)

**TDD Approach**:
- Domain entities: Write validation tests first, then entity implementation
- Use cases: Write business logic tests first, then use case implementation
- Handlers: Write contract tests first, then handler implementation
- Integration: Write scenario tests first, then end-to-end validation

**Constitutional Compliance Tasks**:
- All entity tasks include 80%+ test coverage validation
- All handler tasks include OWASP Top 10 security compliance
- All performance tasks validate 1000+ concurrent users capability
- All migration tasks ensure zero-downtime deployment compatibility

**Estimated Output**: 69 numbered, dependency-ordered tasks in tasks.md with clear TDD progression

**Dependencies Required for /tasks execution**:
- Phase 0: research.md (complete)
- Phase 1: data-model.md, contracts/, quickstart.md (complete)
- Agent context updated for GitHub Copilot (complete)

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
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
