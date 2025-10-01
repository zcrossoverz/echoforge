# Echoforge Documentation

Welcome to the Echoforge project documentation. This section contains detailed information about the architecture, setup, and development workflows.

## 📚 Documentation Structure

### Getting Started
- [Quick Start Guide](quickstart.md) - Get up and running in 5 minutes
- [Installation Guide](installation.md) - Detailed setup instructions
- [Configuration Guide](configuration.md) - Configuration options and examples

### Architecture
- [Hexagonal Architecture](architecture/hexagonal.md) - Ports & adapters pattern
- [Multi-tenancy](architecture/multi-tenancy.md) - Site isolation with `site_id`
- [Domain-Driven Design](architecture/ddd.md) - Domain modeling principles

### Development
- [Development Setup](development/setup.md) - Local development environment
- [Testing Strategy](development/testing.md) - TDD approach and test categories
- [Code Standards](development/standards.md) - Go coding conventions
- [Contributing Guide](development/contributing.md) - How to contribute

### API Documentation
- [REST API Reference](api/rest.md) - HTTP endpoints and examples
- [Authentication](api/auth.md) - JWT token management
- [Error Handling](api/errors.md) - Error response formats

### Operations
- [Deployment Guide](operations/deployment.md) - Production deployment
- [Monitoring](operations/monitoring.md) - Health checks and metrics
- [Security](operations/security.md) - Security best practices
- [Performance](operations/performance.md) - Performance optimization

### Database
- [Schema Design](database/schema.md) - Database structure
- [Migrations](database/migrations.md) - Schema versioning
- [Multi-tenancy](database/multi-tenancy.md) - Site isolation patterns

## 📖 Quick Reference

### Project Structure
```
echoforge/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/         # Business entities (pure)
│   ├── usecase/        # Business logic
│   └── adapters/       # External interfaces
├── pkg/                # Shared packages
├── configs/            # Configuration templates
├── tests/              # Test suites
├── migrations/         # Database migrations
└── docs/              # This documentation
```

### Key Technologies
- **Go 1.25+**: Core language
- **Gin v1.11.0**: HTTP framework
- **GORM v1.31.0**: Database ORM
- **Viper v1.19.0**: Configuration management
- **Zap v1.27.0**: Structured logging
- **Wire v0.7.0**: Dependency injection

### Essential Commands
```bash
# Development
go run cmd/server/main.go

# Testing
go test ./...
go test -cover ./...

# Building
go build -o echoforge cmd/server/main.go

# Optimized build
go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go
```

### Configuration
The main configuration file is `configs/config.yaml`. Key settings:

- `app.site_id`: Multi-tenant site identifier
- `server.port`: HTTP server port (default: 8080)
- `database.url`: PostgreSQL connection string
- `jwt.secret`: JWT token signing secret

## 🔗 External Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)

## 📞 Support

For questions, issues, or contributions:

- **GitHub Issues**: Report bugs and request features
- **GitHub Discussions**: Ask questions and share ideas
- **Code Reviews**: All changes go through pull request review

---

*This documentation is generated automatically and updated with each release.*