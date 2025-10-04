# Echoforge

> A reusable Golang backend core for multi-site content platforms (blog/manga/news) with hexagonal architecture.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-green.svg)](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software))
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 📋 Overview

Echoforge is a modular monolith backend service designed to power multiple content-focused websites. Built with Go 1.25+ and following hexagonal (ports & adapters) architecture principles, it provides a lean, scalable foundation for multi-tenant applications.

### 🎯 Key Features

- **Modular Monolith**: Clean separation between domain, use cases, and adapters
- **Clone-and-Extend**: Each site runs as separate instance with dedicated database (v2.0+ refactor)
- **Hexagonal Architecture**: Domain-driven design with ports & adapters pattern
- **Lean Binary**: Optimized builds under 20MB
- **TDD Approach**: 80%+ test coverage requirement
- **Constitutional Compliance**: SemVer, reproducible builds, zero-downtime deployments

## 🛠 Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| **Language** | Go | 1.25+ |
| **HTTP Framework** | Gin | v1.11.0 |
| **Database ORM** | GORM | v1.31.0 |
| **Database Driver** | PostgreSQL | v1.6.0 |
| **Configuration** | Viper | v1.19.0 |
| **Logging** | Zap | v1.27.0 |
| **Dependency Injection** | Wire | v0.7.0 |
| **Validation** | Validator/v10 | v10.27.0 |
| **Testing** | Testify | v1.11.1 |
| **UUID Generation** | Google UUID | v1.6.0 |
| **Cryptography** | Go Crypto | v0.42.0 |

## 🏗 Project Structure

```
echoforge/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── domain/          # Business entities (pure, no dependencies)
│   ├── usecase/         # Business logic with DI
│   └── adapters/        # External concerns (HTTP, DB, etc.)
│       ├── http/        # Gin HTTP handlers
│       ├── persistence/ # GORM repositories
│       └── logger/      # Zap logging adapter
├── pkg/
│   ├── auth/           # JWT, bcrypt utilities
│   └── common/         # Shared utilities (logger config)
├── configs/            # Site configuration (clone-and-extend model)
├── tests/              # TDD tests (contract, integration, unit)
├── migrations/         # Database schema migrations
└── docs/              # Project documentation
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 16+ (optional for development)
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/zcrossoverz/echoforge.git
   cd echoforge
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

4. **Test the health endpoint**
   ```bash
   curl http://localhost:8080/health
   ```

### Configuration

Create a `configs/config.yaml` file:

```yaml
port: "8080"
log_level: "info"
site_id: "your-site-name"  # For operational identification only
mode: "development"

database:
  dsn: "host=localhost user=echoforge password=echoforge dbname=echoforge port=5432 sslmode=disable"
```

## 🧪 Testing

Run all tests with coverage:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test categories
go test ./tests/contract/...     # Contract tests
go test ./tests/integration/...  # Integration tests
go test ./tests/unit/...         # Unit tests
```

### Test-Driven Development (TDD)

This project follows strict TDD principles:

1. **Red**: Write failing tests first
2. **Green**: Write minimal code to pass tests
3. **Refactor**: Improve code while maintaining tests

## 📦 Building

### Development Build

```bash
go build -o echoforge cmd/server/main.go
```

### Production Build (Optimized)

```bash
go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go
```

### Cross-Platform Builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o echoforge-linux-amd64 cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o echoforge-darwin-amd64 cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o echoforge-windows-amd64.exe cmd/server/main.go
```

## 🐳 Docker

Build and run with Docker:

```dockerfile
# Dockerfile (example)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/echoforge .
COPY --from=builder /app/configs ./configs
CMD ["./echoforge"]
```

## 🏛 Architecture Principles

### Hexagonal Architecture (Ports & Adapters)

- **Domain Layer** (`internal/domain/`): Pure business entities with no external dependencies
- **Use Case Layer** (`internal/usecase/`): Business logic orchestration with dependency injection
- **Adapter Layer** (`internal/adapters/`): External concerns (HTTP, database, logging)

### Clone-and-Extend Architecture

Each site clones the core repository and runs with its own database:

```go
// Example: Simple user query (no site filtering needed)
users := []User{}
db.Find(&users)  // Each site has its own database
```

**Benefits**:
- Natural data isolation per site
- Simplified queries (no site_id joins)  
- Independent scaling and deployment
- Core updates via `go get` (SemVer)

## 📊 Performance Targets

- **Concurrent Users**: 1000+ per site
- **Binary Size**: <20MB unoptimized, <15MB optimized
- **Build Time**: <10 seconds for basic build
- **Memory Usage**: <100MB baseline per site
- **Response Time**: <100ms for typical API calls

## 🔒 Security

- **OWASP Top 10 Compliance**: Input validation, authentication, authorization
- **bcrypt Password Hashing**: Default cost for production security
- **JWT Tokens**: Stateless authentication with configurable expiration
- **Rate Limiting**: Built-in protection against abuse
- **Input Validation**: Comprehensive validation using go-playground/validator

## 🚀 Deployment

### Zero-Downtime Deployment

1. **Blue-Green Deployment**: Run two identical production environments
2. **Health Checks**: `/health` endpoint for load balancer monitoring
3. **Graceful Shutdown**: SIGTERM handling with configurable timeout
4. **Rolling Updates**: Docker container orchestration support

### Environment Variables

```bash
export PORT=8080
export LOG_LEVEL=info
export SITE_ID=my-blog-site  # Operational identifier
export DATABASE_DSN="host=db user=echoforge password=secure dbname=echoforge sslmode=require"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests first (TDD approach)
4. Implement the feature
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Maintain 80%+ test coverage
- Use `gofmt` for code formatting
- Follow hexagonal architecture patterns
- Write clear, self-documenting code

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/zcrossoverz/echoforge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/zcrossoverz/echoforge/discussions)

## 🗺 Roadmap

- [ ] Authentication & Authorization MVP
- [ ] Database Migration System
- [ ] Admin Dashboard API
- [ ] Content Management System
- [ ] Multi-language Support
- [ ] Caching Layer (Redis)
- [ ] Message Queue Integration
- [ ] Monitoring & Observability

---

**Built with ❤️ using Go and hexagonal architecture principles.**