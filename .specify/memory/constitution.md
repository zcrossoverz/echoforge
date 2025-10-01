
<!--
Sync Impact Report
Version change: 1.0.0 → 1.1.0
Modified principles: 
- I. Hexagonal Architecture → expanded with modular monolith guidance
- II. GORM for Postgres → expanded with migration requirements
- IV. TDD → expanded with 80%+ coverage requirement
- V. Multi-Site Tenant Isolation → expanded with reusable core guidance
- VI. Authentication → expanded with unique email constraint
Added sections: 
- VII. Reusable Core Design
- VIII. Semantic Versioning & Backward Compatibility
- IX. Performance & Scalability
- X. Security-First Development
- XI. Lean MVP Approach
Removed sections: None
Templates requiring updates:
✅ .specify/templates/plan-template.md
✅ .specify/templates/spec-template.md
✅ .specify/templates/tasks-template.md
✅ .specify/templates/agent-file-template.md
Follow-up TODOs: TODO(RATIFICATION_DATE): Original adoption date unknown, please specify.
-->

# Echoforge Constitution

## Core Principles

### I. Modular Monolith with Hexagonal Architecture
All business logic MUST be implemented as a modular monolith using hexagonal (ports & adapters) pattern. Core domain entities and interfaces reside in `internal/domain` (pure, framework-agnostic). Use cases implement business logic in `internal/usecase` with dependency injection. Adapters handle external concerns: GORM for persistence, Gin for HTTP, Zap for logging. Clean boundaries prepare for future refactoring to microservices when needed.

### II. GORM for Postgres with Migrations
All data persistence MUST use GORM v1.26+ as the ORM layer, targeting PostgreSQL 16+ as the primary database. Database migrations MUST use golang-migrate for version control. Direct SQL or alternative ORMs are prohibited unless justified and approved via governance. This ensures consistency, migration support, and leverages GORM's ecosystem.

### III. Gin for API Layer
All HTTP APIs MUST be implemented using the Gin v1.10+ framework with versioned endpoints (`/api/v1/`). Middleware, routing, and request/response handling are to be managed via Gin. This provides performance, community support, and a consistent developer experience for 1000+ concurrent users per site.

### IV. Test-Driven Development with 80%+ Coverage
All new features and bug fixes MUST follow TDD using Testify: write failing tests, then implement code to pass. Test coverage MUST exceed 80%. No code is merged without corresponding tests. This enforces reliability, prevents regressions, and ensures lean MVP quality with 500-1000 LOC.

### V. Multi-Site Tenant Isolation via Site ID
All data access and business logic MUST enforce tenant isolation using a `site_id` configuration in both config files and database queries. No cross-tenant data leakage is permitted. This enables reusable core for 10+ sites (blog, truyện, etc.) with proper isolation for security and regulatory compliance.

### VI. Scalable Authentication with Unique Email Constraint
Authentication MUST use bcrypt for password hashing and JWT for stateless session management. Email addresses MUST be unique across the system. All auth endpoints MUST include rate limiting for security. Auth flows must support 1000+ concurrent users per site with horizontal scaling capabilities.

### VII. Reusable Core Design
The core system MUST be designed for reusability across multiple sites through configuration overrides. Sites clone the repository and customize via `config.yaml` (DB_URL, site_id, custom settings) without modifying core logic. Use Factory pattern for multi-site configuration and Repository pattern for data abstraction. Observer pattern for events (e.g., new user registration → notifications).

### VIII. Semantic Versioning & Backward Compatibility
All releases MUST follow Semantic Versioning (SemVer): MAJOR for breaking changes, MINOR for new features, PATCH for bug fixes. Backward compatibility MUST be maintained for API endpoints and configuration. Breaking changes require explicit justification, migration guides, and MAJOR version bump.

### IX. Performance & Scalability
The system MUST handle 1000+ concurrent users per site using Go's concurrency features. Zero-downtime deployments are required using Docker containers with rolling updates. Database queries MUST be optimized for multi-tenant workloads with proper indexing on `site_id`.

### X. Security-First Development
All development MUST comply with OWASP Top 10 security standards. Authentication endpoints MUST implement rate limiting. All user input MUST be validated and sanitized. Security reviews are mandatory for auth-related changes and multi-tenant features.

### XI. Lean MVP Approach
Initial implementation MUST remain lean (500-1000 LOC) following YAGNI principles. No microservices until proven necessary. Focus on core MVP features: user registration, authentication, and multi-site configuration. Complex features require explicit justification.

## Technology Stack

**Core Technologies:**
- Language: Go 1.25+
- API Framework: Gin v1.10+
- ORM: GORM v1.26+
- Database: PostgreSQL 16+
- Logging: Zap v1.27+
- Configuration: Viper (YAML/env)
- Testing: Testify (TDD, 80%+ coverage)
- Dependency Injection: Wire
- Migrations: golang-migrate
- Deployment: Docker with rolling updates

**Architecture Pattern:** Modular Monolith with Hexagonal (Ports & Adapters)

All dependencies and versions MUST be documented in `go.mod` and reviewed for security and compatibility. Version constraints MUST be specified to prevent breaking changes.

## Development Workflow

- All code changes require peer review and MUST pass all tests (80%+ coverage) before merge.
- TDD is enforced: tests precede implementation (Red-Green-Refactor cycle).
- Feature branches follow the `[###-feature-name]` convention.
- All deployments MUST use configuration files for environment-specific settings, including `site_id` for tenant isolation.
- Security reviews are mandatory for all auth-related changes and multi-tenant features.
- Code reviews MUST verify SemVer compliance and backward compatibility.
- Performance testing required for changes affecting concurrency or database queries.

## Governance

This constitution supersedes all other development practices. Amendments require documentation, peer approval, and a migration plan for any breaking changes. All PRs and reviews MUST verify compliance with every principle herein, especially:

- Hexagonal architecture boundaries maintained
- Multi-site tenant isolation enforced
- TDD with 80%+ test coverage achieved
- SemVer compliance verified
- Security standards met (OWASP Top 10)
- Performance requirements satisfied (1000+ concurrent users/site)

Versioning follows semantic rules: MAJOR for breaking/removal, MINOR for new/expanded principles, PATCH for clarifications. Compliance reviews are conducted quarterly or upon major release.

For architecture decisions not covered by this constitution, prefer:
1. Go idioms and community best practices
2. Simplicity over complexity (YAGNI)
3. Testability and maintainability
4. Performance and security

Refer to `README.md` and `.specify/templates/agent-file-template.md` for runtime and development guidance.

**Version**: 1.1.0 | **Ratified**: TODO(RATIFICATION_DATE): Original adoption date unknown, please specify. | **Last Amended**: 2025-10-01