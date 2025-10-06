# Tasks: Documentation Integration and Site Extension Guide

**Input**: Design documents from `/specs/008-write-doc-integrate/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Found: Documentation system with Go tooling and static generation
   → Extract: Go 1.25+, Mermaid diagrams, OpenAPI 3.0, Postman collections
2. Load optional design documents:
   → data-model.md: Site guides, architecture models, API specs, Postman collections
   → contracts/: Documentation generation and validation API endpoints
   → research.md: Mermaid for diagrams, OpenAPI 3.0, automated validation
3. Generate tasks by category:
   → Setup: Directory structure, Go scripts, documentation tools
   → Tests: Documentation validation, link checking, example testing
   → Core: Site guides, architecture diagrams, API documentation
   → Integration: Postman collections, OpenAPI generation, validation APIs
   → Polish: Automation, CI integration, troubleshooting guides
4. Apply task rules:
   → Different documentation sections = mark [P] for parallel
   → Shared tools/scripts = sequential (no [P])
   → Validation tests before content generation
   → All examples must be testable and validated
   → Visual diagrams must render correctly across platforms
   → API documentation must be complete and accurate
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All site types have guides?
   → All diagrams have sources and validation?
   → All API endpoints documented?
   → Postman collections complete?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
Documentation-focused structure with Go tooling:
- **docs/**: Main documentation directory
- **scripts/**: Go-based generation and validation tools
- **tests/**: Documentation validation and testing

## Phase 3.1: Setup
- [x] T001 Create documentation directory structure at docs/guides/, docs/architecture/, docs/api/, docs/postman/
- [x] T002 Create scripts directory with Go modules at scripts/generate-docs.go, scripts/validate-examples.go, scripts/build-postman.go
- [x] T003 [P] Initialize Go module for documentation tools with dependencies: Mermaid CLI, OpenAPI generators, JSON validation

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T004 [P] Documentation validation test in tests/documentation/guide_validation_test.go
- [x] T005 [P] Link checking test in tests/documentation/test_link_validation.go  
- [x] T006 [P] Code example execution test in tests/documentation/test_example_validation.go
- [x] T007 [P] Diagram rendering test in tests/documentation/test_diagram_validation.go
- [x] T008 [P] Postman collection validation test in tests/documentation/test_postman_validation.go
- [x] T009 [P] API documentation completeness test in tests/documentation/test_api_completeness.go

## Phase 3.3: Core Documentation Content (ONLY after tests are failing)
- [ ] T010 [P] Manga site extension guide at docs/guides/site-extension/manga-site-setup.md
- [ ] T011 [P] Blog site extension guide at docs/guides/site-extension/blog-site-setup.md  
- [ ] T012 [P] Portfolio site extension guide at docs/guides/site-extension/portfolio-site-setup.md
- [ ] T013 [P] Hexagonal architecture diagram source at docs/architecture/diagrams/hexagonal-architecture.mmd
- [ ] T014 [P] Multi-tenant data flow diagram at docs/architecture/diagrams/data-flow.mmd
- [ ] T015 [P] Deployment architecture diagram at docs/architecture/diagrams/deployment.mmd
- [ ] T016 Site configuration templates at docs/site-configs/ (manga-site.yaml, blog-site.yaml, portfolio-site.yaml)
- [ ] T017 Customization patterns documentation at docs/guides/customization/

## Phase 3.4: API Documentation and Integration
- [ ] T018 Generate OpenAPI 3.0 specification at docs/api/openapi.yaml from existing codebase
- [ ] T019 Create interactive API documentation with Swagger UI integration
- [ ] T020 Build Postman collection generator script that creates docs/postman/echoforge-api.json
- [ ] T021 Create environment configurations at docs/postman/environments/ (dev.json, staging.json, prod.json)
- [ ] T022 Documentation generation API endpoints implementation (if required by contracts)

## Phase 3.5: Automation and Validation Tools
- [ ] T023 Documentation generator script at scripts/generate-docs.go (processes all content)
- [ ] T024 Example validation script at scripts/validate-examples.go (tests all code samples)
- [ ] T025 Diagram rendering pipeline at scripts/render-diagrams.go (Mermaid to SVG/PNG)
- [ ] T026 Link checking utility at scripts/check-links.go (validates all references)

## Phase 3.6: Polish and Integration
- [ ] T027 [P] Troubleshooting guide at docs/troubleshooting/common-issues.md
- [ ] T028 [P] Performance guidelines at docs/guides/performance/optimization.md  
- [ ] T029 [P] Security best practices at docs/guides/security/implementation.md
- [ ] T030 CI/CD integration for documentation builds and validation
- [ ] T031 README updates with documentation links and quick start
- [ ] T032 Run complete quickstart validation test from docs/quickstart.md

## Dependencies
- Setup (T001-T003) before all other tasks
- Tests (T004-T009) before content creation (T010-T017)
- Site templates (T016) needed for customization docs (T017)
- OpenAPI spec (T018) before Postman generation (T020-T021)
- Content creation before automation tools (T023-T026)
- All content before troubleshooting and polish (T027-T032)

## Parallel Example
```
# Launch site guide creation together:
Task: "Create manga site extension guide at docs/guides/site-extension/manga-site-setup.md"
Task: "Create blog site extension guide at docs/guides/site-extension/blog-site-setup.md"  
Task: "Create portfolio site extension guide at docs/guides/site-extension/portfolio-site-setup.md"

# Launch diagram creation together:
Task: "Create hexagonal architecture diagram at docs/architecture/diagrams/hexagonal-architecture.mmd"
Task: "Create data flow diagram at docs/architecture/diagrams/data-flow.mmd"
Task: "Create deployment diagram at docs/architecture/diagrams/deployment.mmd"

# Launch validation tests together:
Task: "Documentation validation test in tests/documentation/test_guide_validation.go"
Task: "Link checking test in tests/documentation/test_link_validation.go"
Task: "Code example test in tests/documentation/test_example_validation.go"
```

## Notes
- [P] tasks = different files, no dependencies
- All code examples must be executable and validated
- Diagrams must render on GitHub and in documentation sites
- Postman collections must include authentication flows
- Documentation must be mobile-friendly and accessible
- All links and references must be validated automatically

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
   - Documentation generation API → validation and generation tasks
   - Postman collection API → collection building tasks
   
2. **From Data Model**:
   - Site Extension Guides → individual site type tasks [P]
   - Architecture Models → diagram creation tasks [P]
   - Postman Collections → generation and environment tasks
   
3. **From User Stories**:
   - Developer onboarding → quickstart validation
   - Site customization → extension guide tasks
   - API integration → Postman collection tasks

4. **Ordering**:
   - Setup → Tests → Content → Tools → Polish
   - Content can be parallel when in different files
   - Tools depend on content being complete

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All site types (manga, blog, portfolio) have guide tasks
- [x] All diagram types have creation and validation tasks  
- [x] All validation tests come before content creation
- [x] Parallel tasks truly independent (different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] OpenAPI generation before Postman collection creation
- [x] Content creation before automation tool development