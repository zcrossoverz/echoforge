# Echoforge Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-01

## Active Technologies
- Go (Golang)
- Gin (API)
- GORM (Postgres ORM)
- PostgreSQL
- Testify (testing)
- bcrypt, JWT (auth)

## Project Structure
```
adapters/ (http, persistence)
cmd/server/
configs/
internal/domain/
internal/usecase/
pkg/auth/
pkg/common/
tests/
```

## Commands
- All APIs via Gin
- DB migrations via GORM
- Tests via Testify
- Auth flows use bcrypt+JWT

## Code Style
- Idiomatic Go
- Hexagonal architecture (ports & adapters)
- TDD enforced (Testify)
- Tenant isolation via site_id

## Recent Changes
- Constitution v1.0.0: Hexagonal, Gin, GORM, TDD, tenant isolation, scalable auth

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->