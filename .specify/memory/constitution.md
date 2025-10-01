
<!--
Sync Impact Report
Version change: 0.0.0 → 1.0.0
Modified principles: All (template → concrete for Golang backend)
Added sections: Technology Stack, Development Workflow
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

### I. Hexagonal Architecture
All business logic MUST be implemented in a domain layer, isolated from frameworks and external systems. Adapters for HTTP, persistence, and other I/O are strictly separated. This ensures testability, maintainability, and clear boundaries between core logic and infrastructure.

### II. GORM for Postgres
All data persistence MUST use GORM as the ORM layer, targeting PostgreSQL as the primary database. Direct SQL or alternative ORMs are prohibited unless justified and approved via governance. This ensures consistency, migration support, and leverages GORM's ecosystem.

### III. Gin for API Layer
All HTTP APIs MUST be implemented using the Gin framework. Middleware, routing, and request/response handling are to be managed via Gin. This provides performance, community support, and a consistent developer experience.

### IV. Test-Driven Development (TDD) with Testify
All new features and bug fixes MUST follow TDD: write failing tests with Testify, then implement code to pass. No code is merged without corresponding tests. This enforces reliability and prevents regressions.

### V. Multi-Site Tenant Isolation
All data access and business logic MUST enforce tenant isolation using a `site_id` configuration. No cross-tenant data leakage is permitted. This is critical for security and regulatory compliance in multi-tenant deployments.

### VI. Scalable Authentication (bcrypt + JWT)
Authentication MUST use bcrypt for password hashing and JWT for stateless session management. All auth flows must be scalable and secure, supporting horizontal scaling and token revocation strategies as needed.

## Technology Stack

- Language: Go (Golang)
- API: Gin
- ORM: GORM
- Database: PostgreSQL
- Testing: Testify
- Auth: bcrypt, JWT
- Architecture: Hexagonal (Ports & Adapters)

All dependencies and versions MUST be documented in `go.mod` and reviewed for security and compatibility.

## Development Workflow

- All code changes require peer review and MUST pass all tests before merge.
- TDD is enforced: tests precede implementation.
- Feature branches follow the `[###-feature-name]` convention.
- All deployments MUST use configuration files for environment-specific settings, including `site_id` for tenant isolation.
- Security reviews are mandatory for all auth-related changes.

## Governance

This constitution supersedes all other development practices. Amendments require documentation, peer approval, and a migration plan for any breaking changes. All PRs and reviews MUST verify compliance with every principle herein. Versioning follows semantic rules: MAJOR for breaking/removal, MINOR for new/expanded principles, PATCH for clarifications. Compliance reviews are conducted quarterly or upon major release.

Refer to `README.md` and `.specify/templates/agent-file-template.md` for runtime and development guidance.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): Original adoption date unknown, please specify. | **Last Amended**: 2025-10-01