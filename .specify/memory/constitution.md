<!--
Sync Impact Report
Version change: 1.1.0 → 1.2.0
Modified principles:
- V. Multi-Site Tenant Isolation → Replaced with Clone-and-Extend Model
- VI. Scalable Authentication → Removed unique email constraint across system, kept per DB
- IX. Performance & Scalability → Removed site_id indexing, added clone scalability
Added sections: None
Removed sections: None
Templates requiring updates:
✅ .specify/templates/plan-template.md
✅ .specify/templates/spec-template.md
✅ .specify/templates/tasks-template.md
✅ .specify/templates/agent-file-template.md
Follow-up TODOs: Specify RATIFICATION_DATE (original adoption date).
-->

# Echoforge Constitution

## Core Principles

### I. Modular Monolith with Hexagonal Architecture
All business logic MUST be implemented as a modular monolith using hexagonal (ports & adapters) pattern. Core domain entities and interfaces reside in `internal/domain` (pure, framework-agnostic). Use cases implement business logic in `internal/usecase` with dependency injection (Wire). Adapters handle external concerns: GORM for persistence, Gin for HTTP, Zap for logging. Clean boundaries ensure maintainability and future scalability.

### II. GORM for Postgres with Migrations
All data persistence MUST use GORM v1.26+ as the ORM layer, targeting PostgreSQL 16+ as the primary database. Each site instance MUST use a separate Postgres DB. Database migrations MUST use golang-migrate for version control, with additive changes for zero-downtime deploys. Direct SQL or alternative ORMs are prohibited unless justified.

### III. Gin for API Layer
All HTTP APIs MUST use Gin v1.10+ with versioned endpoints (`/api/v1/`). Middleware, routing, and request/response handling are managed via Gin to support 1000+ concurrent users per site with high performance.

### IV. Test-Driven Development with 80%+ Coverage
All features and bug fixes MUST follow TDD using Testify: write failing tests, then implement code. Test coverage MUST exceed 80%. No code merges without tests, ensuring reliability and lean MVP quality (500-1000 LOC).

### V. Clone-and-Extend Model for Multi-Site
Each site (blog/manga/news) MUST clone the core repository (github.com/zcrossoverz/echoforge) and extend via custom configs (Viper) and site-specific features. Each site uses a separate Postgres DB (configured via DB_DSN) for data isolation, eliminating the need for site_id. Core remains generic and reusable, with updates applied via `go get` (SemVer).

### VI. Scalable Authentication
Authentication MUST use bcrypt for password hashing (cost=12) and JWT (HS256) for stateless sessions. Email addresses MUST be unique per site DB. Auth endpoints MUST support rate limiting (anti-brute force) and handle 1000+ concurrent users per site.

### VII. Reusable Core Design
Core MUST be reusable as a Go module (github.com/zcrossoverz/echoforge). Sites customize via `config.yaml` or env (DB_DSN, JWT_SECRET, custom settings) without modifying core logic. Use Factory pattern for entities, Repository pattern for data, and Observer pattern for events (e.g., user signup notifications).

### VIII. Semantic Versioning & Backward Compatibility
Releases MUST follow SemVer: MAJOR for breaking changes, MINOR for new features, PATCH for bug fixes. Backward compatibility is required for APIs and configs. Breaking changes need migration guides and MAJOR bumps.

### IX. Performance & Scalability
System MUST handle 1000+ concurrent users per site using Go's concurrency (goroutines, channels). Zero-downtime deployments are required via Docker with rolling updates. DB queries MUST be optimized with proper indexing (e.g., on email for auth).

### X. Security-First Development
Development MUST comply with OWASP Top 10. Auth endpoints MUST implement rate limiting and input validation (go-playground/validator/v10). No hard-coded secrets. Security reviews are mandatory for auth and DB changes.

### XI. Lean MVP Approach
Initial implementation MUST be lean (500-1000 LOC), following YAGNI. Focus on MVP: user auth (register/login), health check. Complex features (e.g., content, categories) require justification and are deferred to site-specific extensions.

## Technology Stack
- Language: Go 1.25+
- API Framework: Gin v1.10+
- ORM: GORM v1.26+
- Database: PostgreSQL 16+ (separate per site)
- Logging: Zap v1.27+
- Configuration: Viper (YAML/env)
- Testing: Testify (TDD, 80%+ coverage)
- Dependency Injection: Wire
- Migrations: golang-migrate
- Deployment: Docker with rolling updates

**Architecture Pattern**: Modular Monolith with Hexagonal (Ports & Adapters)

Dependencies MUST be pinned in `go.mod`, reviewed for security (govulncheck), and follow SemVer.

## Development Workflow
- Code changes require peer review and 80%+ test coverage.
- TDD: Red-Green-Refactor cycle.
- Feature branches: `[###-feature-name]` convention.
- Deployments use Viper configs (DB_DSN per site).
- Security reviews for auth-related changes.
- SemVer compliance and backward compatibility checks.
- Performance testing for concurrency and DB queries.

## Governance
Constitution supersedes other practices. Amendments need documentation, peer approval, and migration plans. PRs MUST verify:
- Hexagonal boundaries
- DB isolation per site
- TDD with 80%+ coverage
- SemVer compliance
- OWASP Top 10 compliance
- Performance for 1000+ users/site

Versioning: MAJOR for breaks, MINOR for additions, PATCH for clarifications. Compliance reviews quarterly or on major releases.

**Version**: 1.2.0 | **Ratified**: TODO(RATIFICATION_DATE) | **Last Amended**: 2025-10-04