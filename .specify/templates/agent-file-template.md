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
- Constitution v1.1.0: Expanded with modular monolith, performance targets, security requirements, lean MVP principles

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->