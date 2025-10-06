# Tasks: Abstract Post System

**Input**: Design documents from `/specs/007-abstract-post-feature/`
**Prerequisites**: plan.md (complete), research.md (complete), data-model.md (complete), contracts/ (complete)

## Execution Flow (main)
```
1. Load plan.md from feature directory ✓
   → Found: Go 1.25+, GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, Testify v1.11+
   → Structure: Single modular monolith with hexagonal architecture
2. Load design documents ✓
   → data-model.md: 7 entities (Post, PostType, PostCategory, PostTag, PostAttachment, PostVersion, PostMetadata)
   → contracts/: REST API with 25+ endpoints across 6 categories
   → research.md: Technical decisions and architecture patterns
3. Generate tasks by category ✓
   → Setup: Dependencies, migrations, DI setup
   → Tests: TDD with contract tests, integration tests per quickstart scenario
   → Core: 7 domain entities, 3 use case services, repository interfaces
   → Adapters: GORM persistence, Gin HTTP handlers, middleware
   → Security: JWT integration, validation middleware, rate limiting
   → Multi-tenant: site_id isolation integration
   → Performance: Indexing, concurrency optimization
   → Polish: Integration tests, documentation
4. Apply task rules ✓
   → Different files = [P] parallel execution
   → Same file = sequential
   → Tests before implementation (TDD)
   → All tasks follow constitutional requirements
5. Number tasks sequentially: T001-T069
6. Generate dependency graph ✓
7. Create parallel execution examples ✓
8. Validate task completeness ✓
   → All 25+ endpoints have contract tests
   → All 7 entities have domain models
   → All 5 quickstart scenarios have integration tests
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: Repository root with modular monolith structure
- Paths follow existing Echoforge project structure

## Phase 3.1: Setup and Dependencies
- [x] T001 Add GORM v1.26+ and dependencies to go.mod (GORM v1.26.12, golang-migrate/migrate/v4)
- [x] T002 Create Wire dependency injection setup in cmd/server/wire.go for post system
- [x] T003 [P] Create database migration 006_create_post_tables.up.sql in migrations/
- [x] T004 [P] Create database migration 006_create_post_tables.down.sql in migrations/
- [x] T005 [P] Create post configuration struct in configs/post_config.go

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests (API Endpoints)
- [x] T006 [P] Contract test POST /api/v1/posts in tests/contract/post_create_test.go
- [x] T007 [P] Contract test GET /api/v1/posts/{id} in tests/contract/post_get_test.go
- [ ] T008 [P] Contract test PUT /api/v1/posts/{id} in tests/contract/post_update_test.go
- [ ] T009 [P] Contract test DELETE /api/v1/posts/{id} in tests/contract/post_delete_test.go
- [ ] T010 [P] Contract test GET /api/v1/posts in tests/contract/post_list_test.go
- [ ] T011 [P] Contract test GET /api/v1/post-types in tests/contract/post_type_list_test.go
- [ ] T012 [P] Contract test GET /api/v1/post-types/{id} in tests/contract/post_type_get_test.go
- [ ] T013 [P] Contract test GET /api/v1/categories in tests/contract/category_list_test.go
- [ ] T014 [P] Contract test POST /api/v1/categories in tests/contract/category_create_test.go
- [ ] T015 [P] Contract test GET /api/v1/tags in tests/contract/tag_list_test.go
- [ ] T016 [P] Contract test POST /api/v1/tags in tests/contract/tag_create_test.go
- [ ] T017 [P] Contract test GET /api/v1/search in tests/contract/search_global_test.go
- [ ] T018 [P] Contract test POST /api/v1/posts/{postId}/attachments in tests/contract/attachment_upload_test.go
- [ ] T019 [P] Contract test GET /api/v1/posts/{postId}/attachments in tests/contract/attachment_list_test.go
- [ ] T020 [P] Contract test POST /api/v1/posts/bulk in tests/contract/bulk_operations_test.go

### Domain Entity Tests
- [x] T021 [P] Domain test for Post entity in tests/unit/domain/post_test.go
- [ ] T022 [P] Domain test for PostType entity in tests/unit/domain/post_type_test.go
- [ ] T023 [P] Domain test for PostCategory entity in tests/unit/domain/post_category_test.go
- [ ] T024 [P] Domain test for PostTag entity in tests/unit/domain/post_tag_test.go
- [ ] T025 [P] Domain test for PostAttachment entity in tests/unit/domain/post_attachment_test.go
- [ ] T026 [P] Domain test for PostVersion entity in tests/unit/domain/post_version_test.go
- [ ] T027 [P] Domain test for PostMetadata entity in tests/unit/domain/post_metadata_test.go

### Use Case Tests
- [ ] T028 [P] Use case test for PostUsecase in tests/unit/usecase/post_usecase_test.go
- [ ] T029 [P] Use case test for PostTypeUsecase in tests/unit/usecase/post_type_usecase_test.go
- [ ] T030 [P] Use case test for PostSearchUsecase in tests/unit/usecase/post_search_usecase_test.go

### Integration Tests (Quickstart Scenarios)
- [x] T031 [P] Integration test for blog site extension (Scenario 1) in tests/integration/blog_extension_test.go
- [ ] T032 [P] Integration test for manga site extension (Scenario 2) in tests/integration/manga_extension_test.go
- [ ] T033 [P] Integration test for multi-type search (Scenario 3) in tests/integration/search_filtering_test.go
- [ ] T034 [P] Integration test for scheduling workflow (Scenario 4) in tests/integration/scheduling_approval_test.go
- [ ] T035 [P] Integration test for bulk operations (Scenario 5) in tests/integration/bulk_operations_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Domain Entities
- [x] T036 [P] Post entity with validation in internal/domain/post.go
- [x] T037 [P] PostType entity with field definitions in internal/domain/post_type.go
- [x] T038 [P] PostCategory entity with hierarchy support in internal/domain/post_category.go
- [x] T039 [P] PostTag entity with usage tracking in internal/domain/post_tag.go ✅ COMPLETE
- [x] T040 [P] PostAttachment entity with file metadata in internal/domain/post_attachment.go ✅ COMPLETE  
- [x] T041 [P] PostVersion entity with change tracking in internal/domain/post_version.go ✅ COMPLETE
- [x] T042 [P] PostMetadata entity with JSONB support in internal/domain/post_metadata.go ✅ COMPLETE

### Repository Interfaces
- [x] T043 [P] PostRepository interface in internal/domain/post_repository.go ✅ COMPLETE
- [x] T044 [P] PostTypeRepository interface in internal/domain/post_type_repository.go ✅ COMPLETE
- [x] T045 [P] PostSearchRepository interface in internal/domain/post_search_repository.go ✅ COMPLETE

### Use Case Layer
- [x] T046 PostUsecase with CRUD operations in internal/usecase/post_usecase.go ✅ COMPLETE
- [x] T047 PostTypeUsecase with type management in internal/usecase/post_type_usecase.go ✅ COMPLETE
- [x] T048 PostSearchUsecase with filtering and search in internal/usecase/post_search_usecase.go ✅ COMPLETE

## Phase 3.4: Adapter Implementation

### GORM Persistence Layer
- [x] T049 [P] PostRepository GORM implementation in adapters/persistence/post_repository.go
- [x] T050 [P] PostTypeRepository GORM implementation in adapters/persistence/post_type_repository.go
- [x] T051 [P] PostSearchRepository GORM implementation in adapters/persistence/post_search_repository.go

### HTTP Handlers (Gin Framework)
- [x] T052 PostHandler with CRUD endpoints in adapters/http/handlers/post_handler.go ✅ COMPLETE
- [x] T053 PostTypeHandler with type management endpoints in adapters/http/handlers/post_type_handler.go ✅ COMPLETE
- [x] T054 CategoryHandler with category management endpoints in adapters/http/handlers/category_handler.go
- [x] T055 TagHandler with tag management endpoints in adapters/http/handlers/tag_handler.go
- [x] T056 SearchHandler with search and filtering endpoints in adapters/http/handlers/search_handler.go
- [x] T057 AttachmentHandler with file upload endpoints in adapters/http/handlers/attachment_handler.go
- [x] T058 BulkHandler with bulk operations endpoints in adapters/http/handlers/bulk_handler.go

## Phase 3.5: Integration and Middleware

### Security and Validation
- [x] T059 Post validation middleware in adapters/http/middleware/post_validation_middleware.go ✅ COMPLETE
- [x] T060 File upload security middleware in adapters/http/middleware/file_security_middleware.go ✅ COMPLETE
- [x] T061 Rate limiting middleware for post operations in adapters/http/middleware/post_rate_limit_middleware.go ✅ COMPLETE

### Route Registration
- [x] T062 Register post routes in adapters/http/router.go (API versioning /api/v1/) ✅ COMPLETE
- [x] T063 Integrate post handlers with existing JWT authentication middleware ✅ COMPLETE

### Database Integration
- [ ] T064 Run post system migrations and seed default data
- [ ] T065 Add post system indexes for performance optimization

## Phase 3.6: Polish and Validation

### Performance and Monitoring
- [ ] T066 [P] Add performance monitoring for post operations in pkg/common/post_metrics.go
- [ ] T067 [P] Optimize database queries for concurrent access (1000+ users target)

### Documentation and Validation
- [ ] T068 [P] Update API documentation with post endpoints in docs/api.md
- [ ] T069 Execute quickstart validation scenarios and performance benchmarks

## Dependencies

### Phase Dependencies
- Setup (T001-T005) must complete before Tests (T006-T035)
- Tests (T006-T035) must complete and FAIL before Core Implementation (T036-T048)
- Core Implementation (T036-T048) before Adapter Implementation (T049-T058)
- Adapter Implementation (T049-T058) before Integration (T059-T065)
- Integration (T059-T065) before Polish (T066-T069)

### Specific Dependencies
- T003, T004 (migrations) before T064 (run migrations)
- T036-T042 (domain entities) before T043-T045 (repository interfaces)
- T043-T045 (repository interfaces) before T046-T048 (use cases)
- T046-T048 (use cases) before T049-T051 (GORM implementations)
- T049-T051 (repositories) before T052-T058 (handlers)
- T052-T058 (handlers) before T062 (route registration)
- T059-T061 (middleware) before T062 (route registration)
- T062 (routes) before T064 (database integration)
- T064 (database) before T069 (validation scenarios)

### Blocking Dependencies
- T046 blocks T052 (PostUsecase → PostHandler)
- T047 blocks T053 (PostTypeUsecase → PostTypeHandler)
- T048 blocks T056 (PostSearchUsecase → SearchHandler)
- T062 blocks T063 (route registration → auth integration)

## Parallel Execution Examples

### Contract Tests (All Independent)
```bash
# Launch T006-T020 together:
Task: "Contract test POST /api/v1/posts in tests/contract/post_create_test.go"
Task: "Contract test GET /api/v1/posts/{id} in tests/contract/post_get_test.go"
Task: "Contract test PUT /api/v1/posts/{id} in tests/contract/post_update_test.go"
Task: "Contract test DELETE /api/v1/posts/{id} in tests/contract/post_delete_test.go"
Task: "Contract test GET /api/v1/posts in tests/contract/post_list_test.go"
# ... (all 15 contract tests can run in parallel)
```

### Domain Entities (All Independent)
```bash
# Launch T036-T042 together:
Task: "Post entity with validation in internal/domain/post.go"
Task: "PostType entity with field definitions in internal/domain/post_type.go"
Task: "PostCategory entity with hierarchy support in internal/domain/post_category.go"
Task: "PostTag entity with usage tracking in internal/domain/post_tag.go"
Task: "PostAttachment entity with file metadata in internal/domain/post_attachment.go"
Task: "PostVersion entity with change tracking in internal/domain/post_version.go"
Task: "PostMetadata entity with JSONB support in internal/domain/post_metadata.go"
```

### GORM Repositories (All Independent)
```bash
# Launch T049-T051 together:
Task: "PostRepository GORM implementation in adapters/persistence/post_repository.go"
Task: "PostTypeRepository GORM implementation in adapters/persistence/post_type_repository.go"
Task: "PostSearchRepository GORM implementation in adapters/persistence/post_search_repository.go"
```

## Constitutional Compliance Validation

### TDD Requirements (80%+ Coverage)
- All test tasks (T006-T035) must be completed and failing before implementation
- Each domain entity (T036-T042) has corresponding test (T021-T027)
- Each use case (T046-T048) has corresponding test (T028-T030)
- All contract tests (T006-T020) validate API endpoints per OpenAPI spec

### Multi-Tenancy (site_id Isolation)
- All repository implementations (T049-T051) must enforce site_id filtering
- All handlers (T052-T058) must extract site_id from JWT or context
- All database queries must include site_id WHERE clauses

### Performance Requirements (1000+ Concurrent Users)
- T065 (indexing) must create indexes per data-model.md specifications
- T067 (optimization) must validate sub-500ms response times
- T069 (validation) must execute performance benchmarks

### Security (OWASP Top 10)
- T059 (validation) must prevent injection attacks and validate input
- T060 (file security) must validate file types and sizes (100MB limit)
- T061 (rate limiting) must implement 1000 req/hour, 10 uploads/hour limits

### API Versioning (SemVer Compliance)
- T062 (routes) must use /api/v1/ versioning pattern
- All handlers must maintain backward compatibility
- Breaking changes require major version increment

## Notes
- [P] tasks = different files, can run in parallel
- Verify all tests fail before implementing (TDD requirement)
- Commit after each task completion
- All tasks must maintain constitutional compliance
- Execute quickstart scenarios (T069) to validate complete functionality

## Task Generation Rules Applied

1. **From Contracts**: 15 endpoint contracts → 15 contract test tasks (T006-T020)
2. **From Data Model**: 7 entities → 7 domain tasks (T036-T042) + 7 test tasks (T021-T027)
3. **From Quickstart**: 5 scenarios → 5 integration test tasks (T031-T035)
4. **From Plan**: Technical stack → setup tasks (T001-T005)
5. **Architecture**: Hexagonal pattern → use cases (T046-T048), repositories (T043-T045, T049-T051), handlers (T052-T058)

## Validation Checklist ✓

- [x] All 15+ endpoint contracts have corresponding tests (T006-T020)
- [x] All 7 entities have domain model tasks (T036-T042)
- [x] All tests come before implementation (Phase 3.2 before 3.3)
- [x] Parallel tasks are truly independent (different files)
- [x] Each task specifies exact file path
- [x] No [P] task modifies same file as another [P] task
- [x] Dependencies properly ordered (setup → tests → core → adapters → integration → polish)
- [x] Constitutional requirements enforced in relevant tasks
- [x] TDD methodology with 80%+ coverage requirement
- [x] Performance targets validated (1000+ concurrent users)