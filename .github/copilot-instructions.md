# Echoforge Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-01

## Active Technologies
- Go 1.25+ (core language)
- Gin v1.10+ (HTTP API with /api/v1/ versioning)
- GORM v1.26+ (Postgres 16+ ORM)
- Zap v1.27+ (structured logging)
- Viper (config YAML/env with site_id)
- Testify (TDD, 80%+ coverage)
- Wire (dependency injection)
- golang-migrate (DB migrations)
- bcrypt, JWT (authentication with rate limiting)
- Docker (zero-downtime rolling deployments)
- Go 1.25+ + GORM v1.26+ (ORM), Testify (testing), UUID package (unique identifiers), bcrypt (password hashing) (003-define-user-domain)
- PostgreSQL 16+ with GORM ORM, golang-migrate for migrations (003-define-user-domain)
- Go 1.25+ + gin v1.10.0, gorm.io/gorm v1.26.12, go-playground/validator/v10 v10.27.0, golang.org/x/crypto v0.42.0, github.com/golang-jwt/jwt/v5, testify v1.13.1 (004-define-register-and)
- PostgreSQL 16+ with GORM ORM, existing user domain/repository from Task 1.2 (004-define-register-and)
- Go 1.25+ + Viper v1.19.0 (config), Zap v1.27.0 (logging), go-playground/validator/v10 (validation) (006-define-config-and)
- Configuration files (YAML) and environment variables, no database storage for this feature (006-define-config-and)
- Go 1.25+ + GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, bcrypt, JWT v5.3+, Testify v1.11+ (006-i-already-create)
- PostgreSQL 16+ database "bloggo" (host: localhost, user: postgres, password: admin) (006-i-already-create)
- Go 1.25+ (constitutional requirement) + GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, Testify v1.11+ (007-abstract-post-feature)
- PostgreSQL 16+ with separate database per site instance (007-abstract-post-feature)

## Project Structure (Modular Monolith)
```
internal/domain/     # Pure entities, interfaces (hexagonal core)
internal/usecase/    # Business logic with DI
adapters/
├── http/           # Gin HTTP handlers
├── persistence/    # GORM repositories
cmd/server/         # Main application entry
configs/            # Multi-site configuration (site_id)
pkg/auth/           # JWT, bcrypt utilities
pkg/common/         # Shared utilities (logger)
tests/              # TDD tests (80%+ coverage)
migrations/         # golang-migrate DB schemas
```

## Commands & Patterns
- APIs: Gin with middleware, versioned endpoints (/api/v1/)
- DB: GORM with golang-migrate, site_id isolation
- Tests: TDD with Testify, Red-Green-Refactor cycle
- Auth: bcrypt hashing, JWT tokens, unique email constraint
- DI: Wire for clean dependency injection
- Config: Viper with YAML/env overrides per site
- Deployment: Docker with rolling updates for zero-downtime

## Code Style & Architecture
- Idiomatic Go with proper error handling
- Modular monolith with hexagonal architecture (ports & adapters)
- TDD enforced: tests first, 80%+ coverage mandatory
- Multi-site tenant isolation via site_id in all queries
- Reusable core: config override without core modification
- SemVer compliance: backward compatibility required
- Performance: 1000+ concurrent users/site capability
- Security: OWASP Top 10 compliance, input validation, rate limiting
- Lean MVP: 500-1000 LOC, YAGNI principles

## Recent Changes
- 007-abstract-post-feature: Added Go 1.25+ (constitutional requirement) + GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, Testify v1.11+
- 006-i-already-create: Added Go 1.25+ + GORM v1.26+, Gin v1.10+, Zap v1.27+, Viper v1.19+, bcrypt, JWT v5.3+, Testify v1.11+
- 006-define-config-and: Added Go 1.25+ + Viper v1.19.0 (config), Zap v1.27.0 (logging), go-playground/validator/v10 (validation)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
