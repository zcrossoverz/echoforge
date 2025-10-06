
# Implementation Plan: Documentation Integration and Site Extension Guide

**Branch**: `008-write-doc-integrate` | **Date**: October 6, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-write-doc-integrate/spec.md`

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
Create comprehensive documentation system for Echoforge that enables developers to extend and customize sites for specific purposes (manga, blog, e-commerce). Includes visual architecture diagrams, step-by-step guides, interactive API documentation, and ready-to-use Postman collections. Technical approach focuses on static documentation generation with automated validation and visual modeling tools.

## Technical Context
**Language/Version**: Go 1.25+ (constitutional requirement), Markdown, YAML, JSON  
**Primary Dependencies**: Mermaid diagrams, OpenAPI 3.0, Postman collection format, static site generators  
**Storage**: File system based (docs/, configs/, postman/), existing PostgreSQL schema for API documentation  
**Testing**: Documentation validation, link checking, code example testing with Go testing framework  
**Target Platform**: Cross-platform documentation (web browsers, IDEs, Postman client)
**Project Type**: Documentation system for existing Go modular monolith  
**Performance Goals**: Documentation build time <30 seconds, diagram rendering <5 seconds  
**Constraints**: Clean readable format, visual clarity for complex architecture, mobile-friendly  
**Scale/Scope**: Cover all existing APIs, support 10+ site type examples, modular for future expansion

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **Modular monolith with hexagonal architecture** - Documentation will illustrate existing architecture  
✅ **GORM v1.26+ with golang-migrate** - N/A for documentation, will document existing usage  
✅ **Gin v1.10+ with versioned APIs** - N/A for documentation, will document existing API structure  
✅ **TDD with Testify required, 80%+ test coverage** - Documentation validation tests required  
✅ **Multi-site tenant isolation via `site_id`** - Documentation will explain isolation patterns  
✅ **Auth: bcrypt + JWT with unique email constraint** - N/A for documentation, will document existing auth  
✅ **Reusable core design** - Documentation supports core reusability through clear extension guides  
✅ **SemVer compliance** - Documentation versioning will follow project SemVer  
✅ **Performance: 1000+ concurrent users/site** - Documentation will include performance guidelines  
✅ **Security: OWASP Top 10 compliance** - Documentation will include security best practices  
✅ **Lean MVP: 500-1000 LOC limit** - Documentation generation scripts will be lean and focused  
✅ **All dependencies documented** - Documentation dependencies will be minimal and well-documented  
✅ **Zero-downtime deployments** - Documentation deployment will not affect running services

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
docs/                          # Main documentation directory
├── guides/                    # Site extension and customization guides
│   ├── site-extension/        # Step-by-step site creation guides
│   ├── customization/         # Site-specific customization patterns
│   └── deployment/           # Docker and deployment guides
├── architecture/             # Visual architecture documentation
│   ├── diagrams/             # Mermaid diagram sources
│   ├── models/               # Data model documentation
│   └── flows/                # Process and data flow diagrams
├── api/                      # API documentation
│   ├── openapi.yaml          # Complete OpenAPI 3.0 specification
│   ├── examples/             # API usage examples
│   └── authentication/       # Auth flow documentation
├── postman/                  # Postman collections and environments
│   ├── echoforge-api.json    # Complete API collection
│   └── environments/         # Environment configurations
└── troubleshooting/          # Common issues and solutions

scripts/                      # Documentation generation and validation
├── generate-docs.go          # Documentation generator
├── validate-examples.go      # Code example validator
└── build-postman.go          # Postman collection generator

tests/
├── documentation/            # Documentation validation tests
├── examples/                 # Test generated examples
└── integration/              # Documentation integration tests
```

**Structure Decision**: Documentation-focused structure with automated generation and validation. Leverages existing Go ecosystem for tooling while producing static documentation assets that can be hosted independently or integrated into existing documentation systems.

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
- Generate tasks from Phase 1 design docs (contracts, data-model.md, quickstart.md)
- Documentation structure tasks → directory setup and file creation [P]
- Script development tasks → generation and validation tools [P]
- Content creation tasks → guides, examples, and diagrams
- Integration tasks → Postman collections and API documentation
- Validation tasks → automated testing and link checking

**Ordering Strategy**:
- Foundation first: Directory structure and basic scripts
- Content creation: Documentation before validation
- Integration: API docs before Postman collections
- Validation: Tests and checks after content
- Mark [P] for parallel execution (independent documentation sections)

**Estimated Output**: 20-25 numbered, ordered tasks in tasks.md

**Task Categories**:
1. **Infrastructure** (5-6 tasks): Directory setup, script creation, CI integration
2. **Content Generation** (8-10 tasks): Guides, examples, diagrams, troubleshooting
3. **API Integration** (4-5 tasks): OpenAPI spec, Postman collections, authentication
4. **Validation** (3-4 tasks): Testing, link checking, example validation

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
- [x] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
