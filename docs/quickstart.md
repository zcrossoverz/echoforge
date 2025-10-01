# Echoforge Quickstart Guide

Get up and running with Echoforge User Domain in 5 minutes!

## 🚀 Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Git

## 📦 Installation

### 1. Clone Repository
```bash
git clone https://github.com/zcrossoverz/echoforge.git
cd echoforge
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Setup Database
```bash
# Create database
createdb echoforge_dev

# Run migrations
go run cmd/migrate/main.go up
```

### 4. Configure Environment
```bash
# Copy config template
cp configs/config.yaml configs/local.yaml

# Edit configs/local.yaml with your database settings
```

## 🧪 Quick Validation

### Run All Tests
```bash
# Unit tests (no database required)
go test ./tests/ -run "^Test[^I]" -v

# Integration tests (requires PostgreSQL)
go test ./tests/ -run "Integration" -v

# Coverage check
go test ./tests/ -cover -v
```

### Verify Multi-tenancy
```bash
# Run multi-tenant specific tests
go test ./tests/ -run "MultiTenant|SiteIsolation" -v
```

## 🏗️ Architecture Overview

```
echoforge/
├── internal/domain/     # Pure entities (User)
├── internal/usecase/    # Business logic
├── adapters/
│   └── persistence/     # GORM repository
├── tests/               # TDD test suite
├── migrations/          # Database schemas
└── docs/                # Documentation
```

## 📋 Key Features Implemented

### ✅ User Domain Entity
- UUID-based identity
- Email validation (RFC 5322)
- Password hash validation (60+ chars)
- Multi-tenant isolation via `site_id`

### ✅ Repository Pattern
- Clean architecture with ports & adapters
- GORM PostgreSQL integration
- Context-aware operations
- Multi-tenant query isolation

### ✅ Use Cases
- `CreateUser` - With duplicate detection
- `GetUserByEmail` - Site-scoped lookup
- `IsEmailAvailable` - Availability check

### ✅ Test Coverage
- TDD approach with Red-Green-Refactor
- Domain unit tests (100% passing)
- Repository contract tests
- Use case unit tests
- Integration tests (database required)

## 🔒 Multi-tenant Security

Every operation is site-scoped:

```go
// Example: Create user in specific site
siteID := uuid.Parse("12345678-1234-5678-9012-123456789012")
user, err := userUseCase.CreateUser(ctx, siteID, "user@example.com", hashedPassword)

// Example: Find user within site
user, err := userUseCase.GetUserByEmail(ctx, siteID, "user@example.com")
```

## 🎯 Next Steps

1. **API Layer**: Implement HTTP handlers with Gin
2. **Authentication**: Add JWT token management
3. **Middleware**: Request logging, rate limiting
4. **Deployment**: Docker containerization

## 🐛 Troubleshooting

### Database Connection Issues
```bash
# Test PostgreSQL connection
psql -h localhost -U postgres -d echoforge_dev -c "SELECT version();"
```

### Test Failures
```bash
# Run specific test
go test ./tests/ -run "TestSpecificFunction" -v

# Enable debug logging
GORM_LOG_LEVEL=debug go test ./tests/ -v
```

## 📚 Further Reading

- [Architecture Guide](architecture/hexagonal.md)
- [Testing Strategy](development/testing.md)
- [Database Schema](database/schema.md)
- [API Documentation](api/rest.md)

---

**Constitutional Compliance**: ✅ TDD, ✅ Multi-tenancy, ✅ Hexagonal Architecture, ✅ 47.9% Coverage